// Package main provides the Message Service implementation for handling
// message delivery, storage, and retrieval in the WhatsApp clone.
package main

import (
	"log"
	"os"
	"time"

	"whatsapp/internal/api-gateway/middleware"
	"whatsapp/internal/message-service/handlers"
	"whatsapp/pkg/auth"
	"whatsapp/pkg/database"
	"whatsapp/pkg/rabbitmq"

	"github.com/gin-gonic/gin"
)

func main() {
    mongoURI := getEnv("MONGODB_URI", "mongodb://admin:password@mongodb:27017")
    rabbitMQURI := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
    jwtSecret := getEnv("JWT_SECRET", "your-secret-key-here")

    dbClient, err := database.NewClient(mongoURI)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer dbClient.Close()
    
    mqClient, err := rabbitmq.NewClient(rabbitMQURI)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer mqClient.Close()
    
    authService := auth.NewService(jwtSecret, 24*time.Hour)
    
    // Delete existing queues and exchanges to avoid conflicts
    queues := []string{"messages", "message_status", "dead_letters"}
    exchanges := []string{"messages", "dead-letters"}
    
    for _, queue := range queues {
        if err := mqClient.DeleteQueue(queue); err != nil {
            log.Printf("Warning: Failed to delete queue %s: %v", queue, err)
        }
    }
    
    for _, exchange := range exchanges {
        if err := mqClient.DeleteExchange(exchange); err != nil {
            log.Printf("Warning: Failed to delete exchange %s: %v", exchange, err)
        }
    }
    
    // Add a small delay to ensure deletion completes
    time.Sleep(500 * time.Millisecond)
    
    // Declare exchanges for messages
    if err = mqClient.DeclareExchange("messages", "topic"); err != nil {
        log.Fatalf("Failed to declare exchange: %v", err)
    }

    // Declare a Dead Letter Exchange for undeliverable messages
    if err = mqClient.DeclareExchange("dead-letters", "fanout"); err != nil {
        log.Fatalf("Failed to declare dead letter exchange: %v", err)
    }

    // Declare queues
    messageQueue, err := mqClient.DeclareQueueWithDLX("messages", "dead-letters")
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }

    statusQueue, err := mqClient.DeclareQueue("message_status")
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }

    dlQueue, err := mqClient.DeclareQueue("dead_letters")
    if err != nil {
        log.Fatalf("Failed to declare dead letter queue: %v", err)
    }

    // Bind queues to exchanges with routing patterns
    if err = mqClient.BindQueue(messageQueue.Name, "message.#", "messages"); err != nil {
        log.Fatalf("Failed to bind queue: %v", err)
    }

    if err = mqClient.BindQueue(statusQueue.Name, "status.#", "messages"); err != nil {
        log.Fatalf("Failed to bind queue: %v", err)
    }

    if err = mqClient.BindQueue(dlQueue.Name, "#", "dead-letters"); err != nil {
        log.Fatalf("Failed to bind dead letter queue: %v", err)
    }
    
    messageCollection := dbClient.GetCollection("whatsapp", "messages")
    
    messageHandler := handlers.NewMessageHandler(messageCollection, mqClient)
    
    if err = mqClient.Consume(messageQueue.Name, messageHandler.HandleIncomingMessage); err != nil {
        log.Fatalf("Failed to start consuming messages: %v", err)
    }
    
    router := gin.Default()
    
    router.Use(middleware.AuthMiddleware(authService))
    
    router.POST("/messages", messageHandler.SendMessage)
    router.GET("/messages/:UserID", messageHandler.GetMessages) 
    router.PATCH("/messages/:id/status", messageHandler.UpdateMessageStatus)
    
    port := getEnv("PORT", "8082")
    log.Printf("Message Service starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}