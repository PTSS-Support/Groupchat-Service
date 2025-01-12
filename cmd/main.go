package main

import (
	"Groupchat-Service/internal/config"
	"Groupchat-Service/internal/controllers"
	"Groupchat-Service/internal/database/repositories"
	"Groupchat-Service/internal/middleware"
	"Groupchat-Service/internal/services"
	"Groupchat-Service/internal/util"
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"time"
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
	if err != nil {
		log.Fatalf("Failed to create message repository: %v", err)
	}

	fcmTokenRepo, err := repositories.NewFCMTokenRepository(tableClient)
	if err != nil {
		log.Fatalf("Failed to create FCM token repository: %v", err)
	}

	// Initialize health repository
	healthRepo := repositories.NewHealthRepository(cfg.UserServiceURL, util.NewLoggerFactory())

	// Initialize services
	notificationService, err := services.NewNotificationService(cfg.FirebaseCredentialFile, messageRepo)
	if err != nil {
		log.Fatalf("Failed to create notification service: %v", err)
	}

	validationService := services.NewValidationService(cfg.UserServiceURL)
	messageService := services.NewMessageService(messageRepo, fcmTokenRepo, notificationService, validationService)
	fcmTokenService := services.NewFCMTokenService(fcmTokenRepo)
	healthService := services.NewHealthService(healthRepo, util.NewLoggerFactory())

	// Initialize controllers
	messageController := controllers.NewMessageController(messageService, validationService)
	fcmTokenController := controllers.NewFCMTokenController(fcmTokenService, validationService)
	healthController := controllers.NewHealthController(healthService)

	// Set up router
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add Prometheus middleware BEFORE other middleware
	router.Use(middleware.PrometheusMiddleware())

	// Add JWT middleware
	router.Use(middleware.JWTMiddleware(cfg.JWKSURL))

	// Register routes
	messageController.RegisterRoutes(router)
	fcmTokenController.RegisterRoutes(router)
	healthController.RegisterRoutes(router)
	
	middleware.RegisterMetricsEndpoint(router)

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
