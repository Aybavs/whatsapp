package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the database
type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"password" json:"-"` // Never send password in JSON
	Email        string             `bson:"email" json:"email"`
	FullName     string             `bson:"full_name" json:"full_name"`
	AvatarURL    string             `bson:"avatar_url" json:"avatar_url"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	LastLogin    time.Time          `bson:"last_login,omitempty" json:"last_login,omitempty"`
	Status       string             `bson:"status" json:"status"` // online, offline, away
}

// UserRegistration represents the user registration request
type UserRegistration struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

// UserLogin represents the user login request
type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// ProfileUpdate represents the profile update request
type ProfileUpdate struct {
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status"`
}

// StatusUpdate represents a status update request
type StatusUpdate struct {
	Status string `json:"status" binding:"required"`
}

// StatusResponse represents a status update response
type StatusResponse struct {
    UserID string `json:"UserID" example:"5f8d0f1b9d9d9d9d9d9d9d9d"`
    Status string `json:"status" example:"online"`
}


// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewUser(username, email, password string) (*User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID.Hex(),
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		Status:    u.Status,
	}
}
type ContactRequest struct {
    ContactID string `json:"contact_id" binding:"required"`
}

// SuccessResponse is a generic success response
type SuccessResponse struct {
    Message string `json:"message"`
}