package handlers

import (
	"context"
	"net/http"
	"time"

	"whatsapp/pkg/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GroupHandler handles group-related requests
type GroupHandler struct {
	collection *mongo.Collection
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(db *mongo.Database) *GroupHandler {
	return &GroupHandler{
		collection: db.Collection("groups"),
	}
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var input models.GroupRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ownerObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Validate and parse member IDs
	var memberObjectIDs []primitive.ObjectID
	
	// Add owner to members if not already there
	memberObjectIDs = append(memberObjectIDs, ownerObjectID)
	
	for _, memberID := range input.MemberIDs {
		// Skip if it's the owner (already added)
		if memberID == currentUserID.(string) {
			continue
		}
		
		oid, err := primitive.ObjectIDFromHex(memberID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID: " + memberID})
			return
		}
		memberObjectIDs = append(memberObjectIDs, oid)
	}
	
	// Check minimum members (e.g., at least 3 people total for a group? or 2? Let's say 2 including owner)
	if len(memberObjectIDs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group must have at least 2 members"})
		return
	}

	now := time.Now()
	newGroup := models.Group{
		ID:          primitive.NewObjectID(),
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     ownerObjectID,
		MemberIDs:   memberObjectIDs,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = h.collection.InsertOne(context.Background(), newGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// Convert MemberIDs back to strings for response
	var memberIDs []string
	for _, oid := range newGroup.MemberIDs {
		memberIDs = append(memberIDs, oid.Hex())
	}

	groupResponse := models.GroupResponse{
		ID:          newGroup.ID.Hex(),
		Name:        newGroup.Name,
		Description: newGroup.Description,
		OwnerID:     newGroup.OwnerID.Hex(),
		MemberIDs:   memberIDs,
		CreatedAt:   newGroup.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, groupResponse)
}

// GetUserGroups retrieves groups for the current user
func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	currentUserID, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	filter := bson.M{
		"member_ids": currentUserObjectID,
	}

	cursor, err := h.collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())

	var groups []models.Group
	if err := cursor.All(context.Background(), &groups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse groups"})
		return
	}

	groupResponses := []models.GroupResponse{}
	for _, group := range groups {
		var memberIDs []string
		for _, oid := range group.MemberIDs {
			memberIDs = append(memberIDs, oid.Hex())
		}

		groupResponses = append(groupResponses, models.GroupResponse{
			ID:          group.ID.Hex(),
			Name:        group.Name,
			Description: group.Description,
			OwnerID:     group.OwnerID.Hex(),
			MemberIDs:   memberIDs,
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, groupResponses)
}
