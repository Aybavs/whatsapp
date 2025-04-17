// Package main implements the User Service, handling authentication,
// profile management, and user search functionality.
package main

import (
	"context"
	"log"
	"os"
	"time"
	"whatsapp/internal/api-gateway/middleware" // Use the same middleware as message-service
	"whatsapp/internal/user-service/handlers"
	"whatsapp/pkg/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    mongoURI := os.Getenv("MONGODB_URI")
    if (mongoURI == "") {
        mongoURI = "mongodb://localhost:27017"
    }

    jwtSecret := os.Getenv("JWT_SECRET")
    if (jwtSecret == "") {
        jwtSecret = "your-secret-key-here"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer client.Disconnect(ctx)

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)
    }

    log.Println("Connected to MongoDB!")
    
    router := gin.Default()
    authService := auth.NewService(jwtSecret, 24*time.Hour)

    // Create auth middleware - use the same one as in message-service
    authMiddleware := middleware.AuthMiddleware(authService)

    var mongoDB string
    if dbName := os.Getenv("MONGODB_DATABASE"); dbName != "" {
        mongoDB = dbName
    } else {
        mongoDB = "whatsapp"
    }
    
    db := client.Database(mongoDB)
    userHandler := handlers.NewUserHandler(db, authService)
    
    // Public endpoints (no auth required)
    router.POST("/users/register", userHandler.Register)
    router.POST("/users/login", userHandler.Login)
    
    // Protected endpoints (auth required)
    authRoutes := router.Group("")
    authRoutes.Use(authMiddleware) // Apply middleware to all routes in this group
    {
        authRoutes.GET("/users/search", userHandler.SearchUsers)
        authRoutes.GET("/users/contacts", userHandler.GetUserContacts)
        authRoutes.POST("/users/contacts", userHandler.AddContact)
        authRoutes.DELETE("/users/contacts/:id", userHandler.DeleteContact)
        authRoutes.GET("/users/:id", userHandler.GetProfile)     
        authRoutes.PUT("/users/:id", userHandler.UpdateProfile)
        authRoutes.PATCH("/users/:id/status", userHandler.UpdateStatus)
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    
    log.Printf("User Service starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}