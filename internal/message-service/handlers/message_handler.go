package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"whatsapp/pkg/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	messagesCollection *mongo.Collection
	rabbitMQClient     RabbitMQClient
}

// RabbitMQClient interface for messaging
type RabbitMQClient interface {
	Publish(queue string, data interface{}) error
	PublishToExchange(exchange, routingKey string, data interface{}) error
	Consume(queue string, handler func([]byte) error) error
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messagesCollection *mongo.Collection, rabbitMQClient RabbitMQClient) *MessageHandler {
	return &MessageHandler{
		messagesCollection: messagesCollection,
		rabbitMQClient:     rabbitMQClient,
	}
}

// SendMessage godoc
// @Summary      Send a message
// @Description  Sends a message from one user to another
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        message  body      models.MessageRequest  true  "Message Content"
// @Success      201      {object}  models.MessageResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /messages [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var input models.MessageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	receiverObjectID, err := primitive.ObjectIDFromHex(input.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	now := time.Now()
	senderObjectID, _ := primitive.ObjectIDFromHex(senderID.(string))

	newMessage := models.Message{
		ID:         primitive.NewObjectID(),
		SenderID:   senderObjectID,
		ReceiverID: receiverObjectID,
		Content:    input.Content,
		MediaURL:   input.MediaURL,
		CreatedAt:  now,
		Status:     models.MessageStatusSent,
	}

	_, err = h.messagesCollection.InsertOne(context.Background(), newMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	// Use topic exchange with routing key pattern: message.{receiverId}
	routingKey := fmt.Sprintf("message.%s", newMessage.ReceiverID.Hex())
	err = h.rabbitMQClient.PublishToExchange("messages", routingKey, newMessage)
	if err != nil {
		// Fallback to direct queue publishing if exchange publishing fails
		err = h.rabbitMQClient.Publish("messages", newMessage)
		if err != nil {
			c.JSON(http.StatusCreated, gin.H{"warning": "Message saved but real-time delivery may be delayed"})
			return
		}
	}

	messageResponse := models.MessageResponse{
		ID:         newMessage.ID.Hex(),
		SenderID:   newMessage.SenderID.Hex(),
		ReceiverID: newMessage.ReceiverID.Hex(),
		Content:    newMessage.Content,
		MediaURL:   newMessage.MediaURL,
		CreatedAt:  newMessage.CreatedAt.Format(time.RFC3339),
		Status:     string(newMessage.Status),
	}

	c.JSON(http.StatusCreated, messageResponse)
}

// GetMessageHistory is an alias for GetMessages to maintain compatibility with main.go
func (h *MessageHandler) GetMessageHistory(c *gin.Context) {
	h.GetMessages(c)
}

// GetMessages godoc
// @Summary      Get message history
// @Description  Retrieves message history between two users
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        UserID  path      string  true  "User ID to get conversation with"
// @Param        limit    query     int     false "Limit results"
// @Param        before   query     string  false "Get messages before this timestamp"
// @Success      200      {array}   models.MessageResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
func (h *MessageHandler) GetMessages(c *gin.Context) {
	currentUserID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	otherUserID := c.Param("UserID")

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	otherUserObjectID, err := primitive.ObjectIDFromHex(otherUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit := 50 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil {
			limit = parsedLimit
		}
	}

	filter := bson.M{
		"$or": []bson.M{
			{
				"sender_id":   currentUserObjectID,
				"receiver_id": otherUserObjectID,
			},
			{
				"sender_id":   otherUserObjectID,
				"receiver_id": currentUserObjectID,
			},
		},
	}

	if beforeParam := c.Query("before"); beforeParam != "" {
		beforeTime, err := time.Parse(time.RFC3339, beforeParam)
		if err == nil {
			filter["created_at"] = bson.M{"$lt": beforeTime}
		}
	}

	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := h.messagesCollection.Find(context.Background(), filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())

	var messages []models.Message
	if err := cursor.All(context.Background(), &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse messages"})
		return
	}

	// Initialize with empty array instead of nil
	messageResponses := []models.MessageResponse{}
	for _, message := range messages {
		messageResponses = append(messageResponses, models.MessageResponse{
			ID:         message.ID.Hex(),
			SenderID:   message.SenderID.Hex(),
			ReceiverID: message.ReceiverID.Hex(),
			Content:    message.Content,
			MediaURL:   message.MediaURL,
			CreatedAt:  message.CreatedAt.Format(time.RFC3339),
			Status:     string(message.Status),
		})
	}

	go h.markMessagesAsRead(otherUserObjectID, currentUserObjectID)

	c.JSON(http.StatusOK, messageResponses)
}

// UpdateMessageStatus godoc
// @Summary      Update message status
// @Description  Updates the status of a message (delivered, read)
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                   true  "Message ID"
// @Param        status  body      models.MessageStatusUpdate  true  "Status Update"
// @Success      200     {object}  models.MessageStatusResponse
// @Failure      400     {object}  models.ErrorResponse
// @Failure      401     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Router       /messages/{id}/status [patch]
func (h *MessageHandler) UpdateMessageStatus(c *gin.Context) {
	messageID := c.Param("id")

	currentUserID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.MessageStatusUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var message models.Message
	err = h.messagesCollection.FindOne(context.Background(), bson.M{
		"_id": messageObjectID,
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if message.ReceiverID != currentUserObjectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update status of messages sent to you"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"status":     input.Status,
			"updated_at": time.Now(),
		},
	}

	_, err = h.messagesCollection.UpdateOne(context.Background(), bson.M{"_id": messageObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message status"})
		return
	}

	statusUpdate := models.MessageStatusNotification{
		MessageID: messageID,
		Status:    input.Status,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Publish with routing key pattern: status.{messageId}
	routingKey := fmt.Sprintf("status.%s", messageID)
	err = h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate)
	if err != nil {
		// Fallback to direct queue publishing
		_ = h.rabbitMQClient.Publish("message_status", statusUpdate)
	}

	c.JSON(http.StatusOK, models.MessageStatusResponse{
		MessageID: messageID,
		Status:    input.Status,
	})
}

// Helper function to mark messages as read
func (h *MessageHandler) markMessagesAsRead(senderID, receiverID primitive.ObjectID) {
	filter := bson.M{
		"sender_id":   senderID,
		"receiver_id": receiverID,
		"status":      bson.M{"$ne": models.MessageStatusRead},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.MessageStatusRead,
			"updated_at": time.Now(),
		},
	}

	_, _ = h.messagesCollection.UpdateMany(context.Background(), filter, update)

	// Notify about read status updates via RabbitMQ
	// This is a batch operation so we send a composite update
	routingKey := fmt.Sprintf("status.batch.%s.%s", senderID.Hex(), receiverID.Hex())
	statusUpdate := map[string]interface{}{
		"sender_id":   senderID.Hex(),
		"receiver_id": receiverID.Hex(),
		"status":      models.MessageStatusRead,
		"updated_at":  time.Now().Format(time.RFC3339),
		"type":        "batch",
	}

	_ = h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate)
}

// HandleIncomingMessage processes messages from RabbitMQ
func (h *MessageHandler) HandleIncomingMessage(messageData []byte) error {
	var message models.Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		return err
	}

	if message.Status == models.MessageStatusSent {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		update := bson.M{
			"$set": bson.M{
				"status":     models.MessageStatusDelivered,
				"updated_at": time.Now(),
			},
		}

		_, err := h.messagesCollection.UpdateOne(
			ctx,
			bson.M{"_id": message.ID},
			update,
		)
		if err != nil {
			return err
		}

		// Send delivery notification
		statusUpdate := models.MessageStatusNotification{
			MessageID: message.ID.Hex(),
			Status:    models.MessageStatusDelivered,
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		routingKey := fmt.Sprintf("status.%s", message.ID.Hex())
		_ = h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate)
	}

	return nil
}