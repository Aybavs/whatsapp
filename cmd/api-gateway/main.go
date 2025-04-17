// Package main implements the API Gateway service for routing requests to microservices.

// Swagger documentation annotations
// @title           WhatsApp Clone API
// @version         1.0
// @description     A WhatsApp clone backend API built with Go microservices
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @contact.url     http://www.example.com/support
// @contact.email   support@example.com
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
// @security BearerAuth

package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"whatsapp/internal/api-gateway/handlers"
	"whatsapp/internal/api-gateway/middleware"
	"whatsapp/pkg/auth"
	"whatsapp/pkg/rabbitmq"

	_ "whatsapp/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found, using environment variables")
    }

    router := gin.Default()

    userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8081")
    messageServiceURL := getEnv("MESSAGE_SERVICE_URL", "http://localhost:8082")

    jwtSecret := getEnv("JWT_SECRET", "your-secret-key-here")
    
    expirationHours := 24
    if envExpiration := getEnv("JWT_EXPIRATION_HOURS", "24"); envExpiration != "" {
        if parsed, err := strconv.Atoi(envExpiration); err == nil {
            expirationHours = parsed
        }
    }
    
    authService := auth.NewService(jwtSecret, time.Duration(expirationHours)*time.Hour)
    middleware.SetAuthService(authService)

    // Initialize RabbitMQ client
    rabbitMQURI := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
    mqClient, err := rabbitmq.NewClient(rabbitMQURI)
    if err != nil {
        log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
        // Continue without RabbitMQ - handlers will use HTTP fallback
    } else {
        defer mqClient.Close()
        
        // Declare the messages exchange for WebSockets
        if err := mqClient.DeclareExchange("messages", "topic"); err != nil {
            log.Printf("Warning: Failed to declare exchange: %v", err)
        }
    }

    authHandler := handlers.NewAuthHandler(userServiceURL)
    userHandler := handlers.NewUserHandler(userServiceURL)
    messageHandler := handlers.NewMessageHandler(messageServiceURL)
    // Pass the RabbitMQ client to the WebSocket handler
    wsHandler := handlers.NewWebSocketHandler(messageServiceURL, mqClient, authService)

    api := router.Group("/api")
    {
        // User/Auth endpoints
        api.POST("/users/register", authHandler.Register)
        api.POST("/users/login", authHandler.Login)
        
        api.GET("/users/search", middleware.AuthRequired(), userHandler.SearchUsers)
        api.GET("/users/contacts", middleware.AuthRequired(), userHandler.GetUserContacts)
		api.POST("/users/contacts", middleware.AuthRequired(), userHandler.AddContact)
		api.DELETE("/users/contacts/:id", middleware.AuthRequired(), userHandler.DeleteContact)
		
        api.GET("/users/:id", middleware.AuthRequired(), userHandler.GetUserByID)
        api.PUT("/users/:id", middleware.AuthRequired(), userHandler.UpdateProfile)
        api.PATCH("/users/:id/status", middleware.AuthRequired(), userHandler.UpdateStatus)
        
        // Message endpoints
        api.POST("/messages", middleware.AuthRequired(), messageHandler.SendMessage)
        api.GET("/messages/:UserID", middleware.AuthRequired(), messageHandler.GetMessages)
        api.PATCH("/messages/:id/status", middleware.AuthRequired(), messageHandler.UpdateMessageStatus)
        
        // WebSocket endpoint
        api.GET("/ws", wsHandler.HandleWebSocket)
    }

    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    port := getEnv("PORT", "8080")
    log.Printf("API Gateway starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}