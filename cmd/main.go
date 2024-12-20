package main

import (
	"Groupchat-Service/internal/config"
	"Groupchat-Service/internal/controllers"
	"Groupchat-Service/internal/database/repository"
	"Groupchat-Service/internal/middleware"
	"Groupchat-Service/internal/services"
	fcmclient "Groupchat-Service/pkg"
	"context"
	"database/sql"
	"fmt"
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
	// Create the FCM Client configuration
	fcmConfig := &fcmclient.ClientConfig{
		ServerKey:    cfg.FCM.ServerKey,
		MaxRetries:   cfg.FCM.MaxRetries,
		RetryBackoff: cfg.FCM.RetryBackoff,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		RequestsPerSec: 100,
	}

	// Initialize the FCM client with the configuration
	fcmClient, err := fcmclient.NewFCMClient(fcmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FCM client: %w", err)
	}

	// Create the NotificationService with the FCM client
	notificationService := services.NewNotificationService(fcmClient)

	return notificationService, nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", "postgresql://postgres:your_password@localhost:5432/fcm_microservice?sslmode=disable")
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
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

	// Apply global middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware)

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
