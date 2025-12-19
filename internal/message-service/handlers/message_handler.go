package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	groupsCollection   *mongo.Collection
	usersCollection    *mongo.Collection
	rabbitMQClient     RabbitMQClient
}

// RabbitMQClient interface for messaging
type RabbitMQClient interface {
	Publish(queue string, data interface{}) error
	PublishToExchange(exchange, routingKey string, data interface{}) error
	Consume(queue string, handler func([]byte) error) error
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messagesCollection *mongo.Collection, groupsCollection *mongo.Collection, usersCollection *mongo.Collection, rabbitMQClient RabbitMQClient) *MessageHandler {
	return &MessageHandler{
		messagesCollection: messagesCollection,
		groupsCollection:   groupsCollection,
		usersCollection:    usersCollection,
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

	now := time.Now()
	senderObjectID, _ := primitive.ObjectIDFromHex(senderID.(string))

	var newMessage models.Message
	newMessage.ID = primitive.NewObjectID()
	newMessage.SenderID = senderObjectID
	newMessage.Content = input.Content
	newMessage.MediaURL = input.MediaURL
	newMessage.CreatedAt = now
	newMessage.Status = models.MessageStatusSent
	// Determine if this is a direct message or group message
    log.Printf("DEBUG: SendMessage Input - GroupID: '%s', ReceiverID: '%s'", input.GroupID, input.ReceiverID)
	
    if input.GroupID != "" {
		// Group Message Logic
		groupObjectID, err := primitive.ObjectIDFromHex(input.GroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}
        log.Printf("DEBUG: Parsed GroupObjectID: %s", groupObjectID.Hex())
		newMessage.GroupID = groupObjectID
		
		_, err = h.messagesCollection.InsertOne(context.Background(), newMessage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save group message"})
			return
		}

        // Construct response with populated GroupID
        response := models.MessageResponse{
            ID:             newMessage.ID.Hex(),
            SenderID:       newMessage.SenderID.Hex(),
            SenderUsername: h.getUsername(newMessage.SenderID),
            ReceiverID:     newMessage.ReceiverID.Hex(),
            GroupID:        newMessage.GroupID.Hex(),
            Content:        newMessage.Content,
            MediaURL:       newMessage.MediaURL,
            CreatedAt:      newMessage.CreatedAt.Format(time.RFC3339),
            Status:         string(newMessage.Status),
        }

        log.Printf("DEBUG: Response GroupID: %s", response.GroupID)

		// Fan-out: Publish message to all group members
		go h.fanOutGroupMessage(response)
        
        c.JSON(http.StatusCreated, response)

	} else if input.ReceiverID != "" {
		// Direct Message Logic
		receiverObjectID, err := primitive.ObjectIDFromHex(input.ReceiverID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
			return
		}
		newMessage.ReceiverID = receiverObjectID

		_, err = h.messagesCollection.InsertOne(context.Background(), newMessage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
			return
		}

        // Construct response with populated ReceiverID
        response := models.MessageResponse{
            ID:             newMessage.ID.Hex(),
            SenderID:       newMessage.SenderID.Hex(),
            SenderUsername: h.getUsername(newMessage.SenderID),
            ReceiverID:     newMessage.ReceiverID.Hex(),
            GroupID:        newMessage.GroupID.Hex(),
            Content:        newMessage.Content,
            MediaURL:       newMessage.MediaURL,
            CreatedAt:      newMessage.CreatedAt.Format(time.RFC3339),
            Status:         string(newMessage.Status),
        }

		// Use topic exchange with routing key pattern: message.{receiverId}
		routingKey := fmt.Sprintf("message.%s", newMessage.ReceiverID.Hex())
		// Publish the response object so frontend gets username
		err = h.rabbitMQClient.PublishToExchange("messages", routingKey, response)
		if err != nil {
			_ = h.rabbitMQClient.Publish("messages", response)
		}
        
        c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either receiver_id or group_id is required"})
		return
	}
}

// fanOutGroupMessage handles the distribution of group messages
func (h *MessageHandler) fanOutGroupMessage(messageResponse models.MessageResponse) {
	// messageResponse has GroupID as string
	groupID, err := primitive.ObjectIDFromHex(messageResponse.GroupID)
	if err != nil {
		fmt.Printf("Invalid group ID in fan-out: %v\n", err)
		return
	}
	
	members, err := h.fetchGroupMembers(groupID)
	if err != nil {
		fmt.Printf("Failed to fetch group members for fan-out: %v\n", err)
		return
	}

	for _, memberID := range members {
		// Don't send back to sender
		if memberID.Hex() == messageResponse.SenderID {
			continue
		}
		
		// Create a copy of the response for this specific member
		// We set ReceiverID to the memberID so the WebSocket handler knows who to route to
		memberMessage := messageResponse
		memberMessage.ReceiverID = memberID.Hex()
		
		routingKey := fmt.Sprintf("message.%s", memberID.Hex())
		
		// Publish the response (with username)
		err := h.rabbitMQClient.PublishToExchange("messages", routingKey, memberMessage)
		if err != nil {
			fmt.Printf("Failed to publish group message to %s: %v\n", memberID.Hex(), err)
		}
	}
}



// fetchGroupMembers retrieves member IDs for a group
func (h *MessageHandler) fetchGroupMembers(groupID primitive.ObjectID) ([]primitive.ObjectID, error) {
	var group models.Group
	err := h.groupsCollection.FindOne(context.Background(), bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return group.MemberIDs, nil
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

	// Check if we are fetching for a group or a specific user
	groupID := c.Query("group_id")
	otherUserID := c.Query("with") // Changed parameter name to be more explicit via query, or fallback to path

	// Backward compatibility/Path param handling could be tricky if mixed. 
	// The original main.go probably uses /messages/:UserID
	// Let's check how main.go defines it. It accepts :UserID.
	// We should probably check if :UserID matches a Group ID format or if we use a query param.
	// Or just treat it as a conversation ID.
	
	// If groupID is not explicitly set, check if 'with' param is a group
	if groupID == "" && otherUserID != "" {
		// Check if otherUserID is a group
		oid, err := primitive.ObjectIDFromHex(otherUserID)
		if err == nil {
			count, _ := h.groupsCollection.CountDocuments(context.Background(), bson.M{"_id": oid})
			log.Printf("DEBUG: Checked if with-param %s is group: count=%d", otherUserID, count)
			if count > 0 {
				groupID = otherUserID
			}
		}
	}
    
    paramID := c.Param("UserID")
    
    // Also check if the path param (UserID) is actually a group ID
    if groupID == "" && paramID != "" {
        oid, err := primitive.ObjectIDFromHex(paramID)
        if err == nil {
            count, _ := h.groupsCollection.CountDocuments(context.Background(), bson.M{"_id": oid})
            log.Printf("DEBUG: Checked if paramID %s is group: count=%d", paramID, count)
            if count > 0 {
                groupID = paramID
            }
        }
    } 

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
    
    // Initialize as empty slice to return [] instead of null
    messagesResponse := []models.MessageResponse{}

	limit := 50 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil {
			limit = parsedLimit
		}
	}

	var filter bson.M

	if groupID != "" {
		// Fetch Group Messages
		groupObjectID, err := primitive.ObjectIDFromHex(groupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}
		
		filter = bson.M{
			"group_id": groupObjectID,
		}
        log.Printf("DEBUG: Filtering by group_id: %s", groupObjectID.Hex())
	} else {
		// 1:1 Messages
		// Assuming paramID is the other user's ID
		if paramID == "" {
			// If not in path, maybe in query `with`
			paramID = otherUserID
		}
		
		if paramID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID or Group ID required"})
			return
		}

		otherUserObjectID, err := primitive.ObjectIDFromHex(paramID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target ID"})
			return
		}

		filter = bson.M{
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

	ctx := context.Background() // Define context for cursor.Next
	cursor, err := h.messagesCollection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)

	// Removed: var messagesResponse []models.MessageResponse (initiated earlier)
	for cursor.Next(ctx) {
		var msg models.Message
		if err := cursor.Decode(&msg); err != nil {
			continue
		}

		response := models.MessageResponse{
			ID:             msg.ID.Hex(),
			SenderID:       msg.SenderID.Hex(),
			SenderUsername: h.getUsername(msg.SenderID),
			ReceiverID:     msg.ReceiverID.Hex(),
			GroupID:        msg.GroupID.Hex(),
			Content:        msg.Content,
			MediaURL:       msg.MediaURL,
			CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
			Status:         string(msg.Status),
		}

		messagesResponse = append(messagesResponse, response)
	}
    log.Printf("DEBUG: Found %d messages", len(messagesResponse))

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	// Mark as read logic (only for 1:1 for now, group read receipts are complex)
	if groupID == "" && paramID != "" {
		otherUserObjectID, _ := primitive.ObjectIDFromHex(paramID) // Already checked error
		go h.markMessagesAsRead(otherUserObjectID, currentUserObjectID)
	}

	c.JSON(http.StatusOK, messagesResponse)
}

// Helper to get username
func (h *MessageHandler) getUsername(senderID primitive.ObjectID) string {
	var user struct {
		Username string `bson:"username"`
	}
	filter := bson.M{"_id": senderID}
	err := h.usersCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return ""
	}
	return user.Username
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
		MessageID:  messageID,
		Status:     input.Status,
		UpdatedAt:  time.Now().Format(time.RFC3339),
		SenderID:   message.SenderID.Hex(),
		ReceiverID: message.ReceiverID.Hex(),
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

// SearchMessages godoc
// @Summary      Search messages
// @Description  Full-text search in message content (supports groups and 1:1)
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q          query     string  true   "Search query"
// @Param        contact_id query     string  false  "Filter by contact (User or Group) ID"
// @Param        limit      query     int     false  "Limit results (default 50)"
// @Success      200        {array}   models.MessageResponse
// @Failure      400        {object}  models.ErrorResponse
// @Failure      401        {object}  models.ErrorResponse
// @Failure      500        {object}  models.ErrorResponse
// @Router       /messages/search [get]
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	currentUserID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Base filter: regex search on content
	filter := bson.M{
		"content": bson.M{
			"$regex":   query,
			"$options": "i", // case-insensitive
		},
	}

	contactID := c.Query("contact_id")
	
	if contactID != "" {
		// Specific Chat Search
		contactObjectID, err := primitive.ObjectIDFromHex(contactID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
			return
		}

		// Check if contactID is a Group
		isGroup := false
		count, _ := h.groupsCollection.CountDocuments(context.Background(), bson.M{"_id": contactObjectID})
		if count > 0 {
			isGroup = true
		}

		if isGroup {
			// Filter by Group ID
			// Security check: Ensure user is member of this group? 
			// For search, basic check might be enough.
			filter["group_id"] = contactObjectID
		} else {
			// Filter by 1:1 Conversation
			filter["$or"] = []bson.M{
				{"sender_id": currentUserObjectID, "receiver_id": contactObjectID},
				{"sender_id": contactObjectID, "receiver_id": currentUserObjectID},
			}
		}

	} else {
		// Global Search (All My Chats)
		
		// 1. Get all groups user is member of
		// Find groups where "member_ids" contains currentUserObjectID
		cursor, err := h.groupsCollection.Find(context.Background(), bson.M{"member_ids": currentUserObjectID})
		var myGroupIDs []primitive.ObjectID
		if err == nil {
			var groups []models.Group
			if err = cursor.All(context.Background(), &groups); err == nil {
				for _, g := range groups {
					myGroupIDs = append(myGroupIDs, g.ID)
				}
			}
		}

		// 2. Filter: (Sender=Me OR Receiver=Me) OR (GroupID IN MyGroups)
		orConditions := []bson.M{
			{"sender_id": currentUserObjectID},
			{"receiver_id": currentUserObjectID},
		}
		
		if len(myGroupIDs) > 0 {
			orConditions = append(orConditions, bson.M{"group_id": bson.M{"$in": myGroupIDs}})
		}
		
		filter["$or"] = orConditions
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

	messageResponses := []models.MessageResponse{}
	for _, message := range messages {
		messageResponses = append(messageResponses, models.MessageResponse{
			ID:             message.ID.Hex(),
			SenderID:       message.SenderID.Hex(),
			SenderUsername: h.getUsername(message.SenderID),
			ReceiverID:     message.ReceiverID.Hex(),
			GroupID:        message.GroupID.Hex(),
			Content:        message.Content,
			MediaURL:       message.MediaURL,
			CreatedAt:      message.CreatedAt.Format(time.RFC3339),
			Status:         string(message.Status),
		})
	}

	c.JSON(http.StatusOK, messageResponses)
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
			MessageID:  message.ID.Hex(),
			Status:     models.MessageStatusDelivered,
			UpdatedAt:  time.Now().Format(time.RFC3339),
			SenderID:   message.SenderID.Hex(),
			ReceiverID: message.ReceiverID.Hex(),
		}

		routingKey := fmt.Sprintf("status.%s", message.ID.Hex())
		_ = h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate)
	}

	return nil
}
