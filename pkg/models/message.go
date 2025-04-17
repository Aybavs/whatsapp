package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// Message represents a message in the database
type Message struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	ReceiverID primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`
	Content    string             `bson:"content" json:"content"`
	MediaURL   string             `bson:"media_url,omitempty" json:"media_url,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	Status     MessageStatus      `bson:"status" json:"status"`
}

// MessageRequest represents a request to send a message
type MessageRequest struct {
	ReceiverID string `json:"receiver_id" example:"5f8d0f1b9d9d9d9d9d9d9d9d" binding:"required"`
	Content    string `json:"content" example:"Hello, how are you?" binding:"required"`
	MediaURL   string `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID         string `json:"id" example:"5f8d0f1b9d9d9d9d9d9d9d9f"`
	SenderID   string `json:"sender_id" example:"5f8d0f1b9d9d9d9d9d9d9d9d"`
	ReceiverID string `json:"receiver_id" example:"5f8d0f1b9d9d9d9d9d9d9d9e"`
	Content    string `json:"content" example:"Hello, how are you?"`
	MediaURL   string `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	CreatedAt  string `json:"created_at" example:"2023-08-01T15:04:05Z"`
	Status     string `json:"status" example:"delivered"`
}

// MessageStatusUpdate represents a request to update message status
type MessageStatusUpdate struct {
	Status MessageStatus `json:"status" example:"read" binding:"required"`
}

// MessageStatusResponse represents a message status update response
type MessageStatusResponse struct {
	MessageID string        `json:"message_id" example:"5f8d0f1b9d9d9d9d9d9d9d9f"`
	Status    MessageStatus `json:"status" example:"read"`
}

// MessageStatusNotification represents a notification about message status change
type MessageStatusNotification struct {
	MessageID string        `json:"message_id" example:"5f8d0f1b9d9d9d9d9d9d9d9f"`
	Status    MessageStatus `json:"status" example:"read"`
	UpdatedAt string        `json:"updated_at" example:"2023-08-01T15:04:05Z"`
}
