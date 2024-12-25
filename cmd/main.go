package main

import (
	"Groupchat-Service/internal/config"
	"Groupchat-Service/internal/controllers"
	"Groupchat-Service/internal/database/repositories"
	"Groupchat-Service/internal/middleware"
	"Groupchat-Service/internal/services"
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize Azure Table client
	tableClient, err := repositories.NewTableClient(cfg.AzureConnectionString)
	if err != nil {
		log.Fatalf("Failed to create table client: %v", err)
	}

	// Initialize repositories
	messageRepo, err := repositories.NewMessageRepository(tableClient)
	fcmTokenRepo, err := repositories.NewFCMTokenRepository(tableClient)

	// Initialize services
	notificationService, err := services.NewNotificationService(cfg.FirebaseCredentialFile)
	messageService := services.NewMessageService(messageRepo, fcmTokenRepo, notificationService)

	// Initialize controllers
	messageController := controllers.NewMessageController(messageService)

	// Set up router
	router := gin.Default()

	// Add Prometheus middleware BEFORE other middleware
	router.Use(middleware.PrometheusMiddleware())

	// Define routes
	router.GET("/groups/:groupId/messages", messageController.GetMessages)
	router.POST("/groups/:groupId/messages", messageController.CreateMessage)
	router.PUT("/groups/:groupId/messages/:messageId/pin", messageController.ToggleMessagePin)

	// Get the path to the serviceAccountKey.json file
	serviceAccountPath := cfg.FirebaseCredentialFile
	if serviceAccountPath == "" {
		log.Fatalf("FIREBASE_CREDENTIAL_FILE is not set correctly")
	}

	// Initialize Firebase with service account
	opt := option.WithCredentialsFile(serviceAccountPath)
	_, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	fmt.Println("Firebase initialized successfully")

	// Start server
	port := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server exited properly")
}
