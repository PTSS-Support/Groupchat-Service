package main

import (
	"Groupchat-Service/internal/config"
	"Groupchat-Service/internal/controllers"
	"Groupchat-Service/internal/database/repository"
	"Groupchat-Service/internal/services"
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	_ "github.com/lib/pq"
)

func initializeServices(cfg *config.Config) (*services.NotificationService, error) {
	// Use Firebase Admin SDK configuration with the server key
	opt := option.WithCredentialsFile("path/to/serviceAccountKey.json") // JSON key file path

	// Initialize Firebase App
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase App: %w", err)
	}

	// Get the messaging client
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase Messaging client: %w", err)
	}

	// Initialize Notification Service
	notificationService := services.NewNotificationService(client)

	return notificationService, nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	notificationService, err := initializeServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Initialize repositories
	messageRepo := repository.NewMessageRepository(db)
	fcmTokenRepo := repository.NewFCMTokenRepository(db)

	// Initialize message service with dependencies
	messageService := services.NewMessageService(
		messageRepo,
		fcmTokenRepo,
		notificationService,
	)

	// Initialize controllers
	groupMessagesController := controllers.NewGroupMessagesController(messageService)

	// Create router and register routes
	router := mux.NewRouter()

	// Register controller routes
	groupMessagesController.RegisterRoutes(router)

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      c.Handler(router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
