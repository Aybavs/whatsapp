package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Group represents a chat group
type Group struct {
	ID          primitive.ObjectID   `bson:"_id" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	OwnerID     primitive.ObjectID   `bson:"owner_id" json:"owner_id"`
	MemberIDs   []primitive.ObjectID `bson:"member_ids" json:"member_ids"`
	AvatarURL   string               `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

// GroupRequest represents a request to create a group
type GroupRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"member_ids" binding:"required"`
}

// GroupResponse represents a group in API responses
type GroupResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	OwnerID     string   `json:"owner_id"`
	MemberIDs   []string `json:"member_ids"`
	AvatarURL   string   `json:"avatar_url,omitempty"`
	CreatedAt   string   `json:"created_at"`
}
