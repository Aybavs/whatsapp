package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"whatsapp/pkg/auth"
	"whatsapp/pkg/models"
)

// UserHandler handles user-related requests
type UserHandler struct {
    usersCollection *mongo.Collection
    authService     *auth.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *mongo.Database, authService *auth.Service) *UserHandler {
    return &UserHandler{
        usersCollection: db.Collection("users"),
        authService:     authService,
    }
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserRegistration  true  "User Registration Info"
// @Success      201   {object}  models.UserResponse
// @Failure      400   {object}  models.ErrorResponse
// @Failure      409   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
    var input models.UserRegistration
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var existingUser models.User
    err := h.usersCollection.FindOne(context.Background(), bson.M{"username": input.Username}).Decode(&existingUser)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        return
    } else if err != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    err = h.usersCollection.FindOne(context.Background(), bson.M{"email": input.Email}).Decode(&existingUser)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
        return
    } else if err != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    now := time.Now()
    newUser := models.User{
        ID:           primitive.NewObjectID(),
        Username:     input.Username,
        PasswordHash: string(hashedPassword),
        Email:        input.Email,
        FullName:     input.FullName,
        AvatarURL:    input.AvatarURL,
        CreatedAt:    now,
        UpdatedAt:    now,
        Status:       "online",
    }

    _, err = h.usersCollection.InsertOne(context.Background(), newUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    userResponse := models.UserResponse{
        ID:        newUser.ID.Hex(),
        Username:  newUser.Username,
        Email:     newUser.Email,
        FullName:  newUser.FullName,
        AvatarURL: newUser.AvatarURL,
        CreatedAt: newUser.CreatedAt.Format(time.RFC3339),
        Status:    newUser.Status,
    }

    c.JSON(http.StatusCreated, userResponse)
}

// Login godoc
// @Summary      Login a user
// @Description  Authenticate a user and returns JWT token
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.UserLogin  true  "User Credentials"
// @Success      200          {object}  models.LoginResponse
// @Failure      400          {object}  models.ErrorResponse
// @Failure      401          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
    var input models.UserLogin
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    err := h.usersCollection.FindOne(context.Background(), bson.M{"username": input.Username}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, expiration, err := h.authService.GenerateToken(user.ID.Hex(), user.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    update := bson.M{
        "$set": bson.M{
            "last_login": time.Now(),
            "status":     "online",
        },
    }
    _, err = h.usersCollection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, update)
    if err != nil {
        log.Printf("Failed to update last login time: %v", err)
    }

    userResponse := models.UserResponse{
        ID:        user.ID.Hex(),
        Username:  user.Username,
        Email:     user.Email,
        FullName:  user.FullName,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt.Format(time.RFC3339),
        Status:    "online",
    }

    c.JSON(http.StatusOK, models.LoginResponse{
        Token:     token,
        ExpiresAt: expiration.Format(time.RFC3339),
        User:      userResponse,
    })
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Retrieves the user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  models.UserResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
    UserID := c.Param("id")

    objectID, err := primitive.ObjectIDFromHex(UserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    var user models.User
    err = h.usersCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    userResponse := models.UserResponse{
        ID:        user.ID.Hex(),
        Username:  user.Username,
        Email:     user.Email,
        FullName:  user.FullName,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt.Format(time.RFC3339),
        Status:    user.Status,
    }

    c.JSON(http.StatusOK, userResponse)
}

// SearchUsers godoc
// @Summary      Search for users
// @Description  Searches for users by username or full name
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        query  query     string  true  "Search query"
// @Param        limit  query     int     false "Limit results"
// @Success      200    {array}   models.UserResponse
// @Failure      400    {object}  models.ErrorResponse
// @Failure      500    {object}  models.ErrorResponse
// @Router       /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
    query := c.Query("query")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
        return
    }

    limit := 10
    if limitParam := c.Query("limit"); limitParam != "" {
        if _, err := json.Number(limitParam).Int64(); err == nil {
            json.Number(limitParam).Int64()
        }
    }

    filter := bson.M{
        "$or": []bson.M{
            {"username": bson.M{"$regex": query, "$options": "i"}},
            {"full_name": bson.M{"$regex": query, "$options": "i"}},
        },
    }

    findOptions := options.Find().
        SetLimit(int64(limit)).
        SetProjection(bson.M{"password": 0})

    cursor, err := h.usersCollection.Find(context.Background(), filter, findOptions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer cursor.Close(context.Background())

    var users []models.User
    if err := cursor.All(context.Background(), &users); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse search results"})
        return
    }

    var userResponses []models.UserResponse
    for _, user := range users {
        userResponses = append(userResponses, models.UserResponse{
            ID:        user.ID.Hex(),
            Username:  user.Username,
            Email:     user.Email,
            FullName:  user.FullName,
            AvatarURL: user.AvatarURL,
            CreatedAt: user.CreatedAt.Format(time.RFC3339),
            Status:    user.Status,
        })
    }

    c.JSON(http.StatusOK, userResponses)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates user profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id      path      string                true  "User ID"
// @Param        profile body      models.ProfileUpdate  true  "Profile Update Info"
// @Success      200     {object}  models.UserResponse
// @Failure      400     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
    UserID := c.Param("id")
    
    tokenUserID, exists := c.Get("UserID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    if tokenUserID != UserID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
        return
    }

    var input models.ProfileUpdate
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    objectID, err := primitive.ObjectIDFromHex(UserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    update := bson.M{
        "$set": bson.M{
            "updated_at": time.Now(),
        },
    }

    updateSet := update["$set"].(bson.M)
    if input.FullName != "" {
        updateSet["full_name"] = input.FullName
    }
    if input.AvatarURL != "" {
        updateSet["avatar_url"] = input.AvatarURL
    }
    if input.Status != "" {
        updateSet["status"] = input.Status
    }

    result, err := h.usersCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    var user models.User
    err = h.usersCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated profile"})
        return
    }

    userResponse := models.UserResponse{
        ID:        user.ID.Hex(),
        Username:  user.Username,
        Email:     user.Email,
        FullName:  user.FullName,
        AvatarURL: user.AvatarURL,
        CreatedAt: user.CreatedAt.Format(time.RFC3339),
        Status:    user.Status,
    }

    c.JSON(http.StatusOK, userResponse)
}

// UpdateStatus godoc
// @Summary      Update user status
// @Description  Updates a user's online status
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id      path      string                 true  "User ID"
// @Param        status  body      models.StatusUpdate    true  "Status Update"
// @Success      200     {object}  models.StatusResponse
// @Failure      400     {object}  models.ErrorResponse
// @Failure      404     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Router       /users/{id}/status [patch]
func (h *UserHandler) UpdateStatus(c *gin.Context) {
    UserID := c.Param("id")
    
    var input models.StatusUpdate
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    objectID, err := primitive.ObjectIDFromHex(UserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    update := bson.M{
        "$set": bson.M{
            "status":     input.Status,
            "updated_at": time.Now(),
        },
    }

    result, err := h.usersCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "UserID": UserID,
        "status":  input.Status,
    })
}
// GetUserContacts godoc
// @Summary      Get user contacts
// @Description  Retrieves the list of users that the current user has exchanged messages with or added as contacts
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.UserResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /users/contacts [get]
func (h *UserHandler) GetUserContacts(c *gin.Context) {
    // Get the user ID from the authentication token
    UserID, exists := c.Get("UserID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    objectID, err := primitive.ObjectIDFromHex(UserID.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    // Set to collect unique contact IDs
    contactIDsMap := make(map[primitive.ObjectID]bool)
    
    // 1. Get contacts from message history
    messagesCollection := h.usersCollection.Database().Collection("messages")
    pipeline := []bson.M{
        {
            "$match": bson.M{
                "$or": []bson.M{
                    {"sender_id": objectID},
                    {"receiver_id": objectID},
                },
            },
        },
        {
            "$project": bson.M{
                "contact_id": bson.M{
                    "$cond": bson.M{
                        "if":   bson.M{"$eq": []interface{}{"$sender_id", objectID}},
                        "then": "$receiver_id",
                        "else": "$sender_id",
                    },
                },
            },
        },
        {
            "$group": bson.M{
                "_id": "$contact_id",
            },
        },
    }

    cursor, err := messagesCollection.Aggregate(context.Background(), pipeline)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contacts from messages"})
        return
    }
    defer cursor.Close(context.Background())

    var results []struct {
        ID primitive.ObjectID `bson:"_id"`
    }
    
    if err := cursor.All(context.Background(), &results); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse contact results"})
        return
    }

    // Add message contacts to the map
    for _, result := range results {
        contactIDsMap[result.ID] = true
    }

    // 2. Get explicitly added contacts
    contactsCollection := h.usersCollection.Database().Collection("contacts")
    contactsCursor, err := contactsCollection.Find(
        context.Background(),
        bson.M{"UserID": objectID},
    )
    
    if err != nil && err != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve added contacts"})
        return
    }

    if err != mongo.ErrNoDocuments {
        defer contactsCursor.Close(context.Background())

        var explicitContacts []struct {
            ContactID primitive.ObjectID `bson:"contact_id"`
        }

        if err := contactsCursor.All(context.Background(), &explicitContacts); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse contacts data"})
            return
        }

        // Add explicit contacts to the map
        for _, contact := range explicitContacts {
            contactIDsMap[contact.ContactID] = true
        }
    }

    // If no contacts found in either source, return empty array
    if len(contactIDsMap) == 0 {
        c.JSON(http.StatusOK, []models.UserResponse{})
        return
    }

    // Convert map keys to array of contact IDs
    var contactIDs []primitive.ObjectID
    for id := range contactIDsMap {
        contactIDs = append(contactIDs, id)
    }

    // Query the users collection to get contact details
    userCursor, err := h.usersCollection.Find(
        context.Background(),
        bson.M{"_id": bson.M{"$in": contactIDs}},
    )
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contact details"})
        return
    }
    defer userCursor.Close(context.Background())

    // Convert users to response format
    var users []models.User
    if err := userCursor.All(context.Background(), &users); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user results"})
        return
    }

    var userResponses []models.UserResponse
    for _, user := range users {
        userResponses = append(userResponses, models.UserResponse{
            ID:        user.ID.Hex(),
            Username:  user.Username,
            Email:     user.Email,
            FullName:  user.FullName,
            AvatarURL: user.AvatarURL,
            CreatedAt: user.CreatedAt.Format(time.RFC3339),
            Status:    user.Status,
        })
    }

    c.JSON(http.StatusOK, userResponses)
}

// AddContact godoc
// @Summary      Add a new contact
// @Description  Adds a user as a contact for the current user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        contact  body     models.ContactRequest  true  "Contact Details"
// @Success      201      {object} models.SuccessResponse
// @Failure      400      {object} models.ErrorResponse
// @Failure      401      {object} models.ErrorResponse 
// @Failure      404      {object} models.ErrorResponse
// @Failure      500      {object} models.ErrorResponse
// @Router       /users/contacts [post]
func (h *UserHandler) AddContact(c *gin.Context) {
    // Get the user ID from the authentication token
    UserID, exists := c.Get("UserID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    var input models.ContactRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate that contactId is provided
    if input.ContactID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Contact ID is required"})
        return
    }

    // Convert string IDs to ObjectID
    userObjectID, err := primitive.ObjectIDFromHex(UserID.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    contactObjectID, err := primitive.ObjectIDFromHex(input.ContactID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
        return
    }

    // Verify the contact exists
    var contactUser models.User
    err = h.usersCollection.FindOne(context.Background(), bson.M{"_id": contactObjectID}).Decode(&contactUser)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Contact user not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    // Connect to the contacts collection (create if doesn't exist)
    contactsCollection := h.usersCollection.Database().Collection("contacts")

    // Check if contact already exists to prevent duplicates
    existingContact := contactsCollection.FindOne(context.Background(), bson.M{
        "UserID":    userObjectID,
        "contact_id": contactObjectID,
    })
    
    if existingContact.Err() == nil {
        // Contact already exists
        c.JSON(http.StatusOK, gin.H{"message": "Contact already exists"})
        return
    } else if existingContact.Err() != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Create the contact relationship
    contactDoc := bson.M{
        "UserID":    userObjectID,
        "contact_id": contactObjectID,
        "created_at": time.Now(),
    }

    _, err = contactsCollection.InsertOne(context.Background(), contactDoc)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add contact"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Contact added successfully",
        "contact_id": input.ContactID,
    })
}

// DeleteContact godoc
// @Summary      Delete a contact
// @Description  Removes a user from the current user's contacts
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Contact ID"
// @Success      200  {object}  models.SuccessResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /users/contacts/{id} [delete]
func (h *UserHandler) DeleteContact(c *gin.Context) {
    // Get the user ID from the authentication token
    UserID, exists := c.Get("UserID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Get the contact ID from the URL parameter
    contactID := c.Param("id")
    if contactID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Contact ID is required"})
        return
    }

    // Convert string IDs to ObjectID
    userObjectID, err := primitive.ObjectIDFromHex(UserID.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    contactObjectID, err := primitive.ObjectIDFromHex(contactID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
        return
    }

    // Connect to the contacts collection
    contactsCollection := h.usersCollection.Database().Collection("contacts")

    // Delete the contact
    result, err := contactsCollection.DeleteOne(context.Background(), bson.M{
        "UserID":    userObjectID,
        "contact_id": contactObjectID,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Contact deleted successfully",
        "contact_id": contactID,
    })
}