package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests by proxying them to the user service
type AuthHandler struct {
    userServiceURL string
}

// NewAuthHandler creates a new AuthHandler instance with the specified user service URL
func NewAuthHandler(userServiceURL string) *AuthHandler {
    return &AuthHandler{
        userServiceURL: userServiceURL,
    }
}

// Register handles user registration requests
func (h *AuthHandler) Register(c *gin.Context) {
    h.proxyRequest(c, "/users/register", http.MethodPost)
}

// Login handles user authentication requests
func (h *AuthHandler) Login(c *gin.Context) {
    h.proxyRequest(c, "/users/login", http.MethodPost)
}

// GetUserByID retrieves a user profile by ID
func (h *AuthHandler) GetUserByID(c *gin.Context) {
    UserID := c.Param("id")
    h.proxyRequest(c, "/users/"+UserID, http.MethodGet)
}

// UpdateProfile handles user profile update requests with permission checking
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
    UserID := c.Param("id")

    tokenUserID, exists := c.Get("UserID")
    if !exists || tokenUserID != UserID {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "You can only update your own profile",
        })
        return
    }

    h.proxyRequest(c, "/users/"+UserID, http.MethodPut)
}

// proxyRequest forwards requests to the user service and handles the response
func (h *AuthHandler) proxyRequest(c *gin.Context, path string, method string) {
    log.Printf("Proxying request to: %s%s with method: %s", h.userServiceURL, path, method)
    
    var requestBody []byte
    if c.Request.Body != nil {
        requestBody, _ = io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        log.Printf("Request body: %s", string(requestBody))
    }

    req, err := http.NewRequest(method, h.userServiceURL+path, bytes.NewBuffer(requestBody))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
        return
    }

    req.Header = c.Request.Header
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User service unavailable"})
        return
    }
    defer resp.Body.Close()

    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
        return
    }

    log.Printf("Response status: %d", resp.StatusCode)
    log.Printf("Response body: %s", string(responseBody))

    for key, values := range resp.Header {
        for _, value := range values {
            c.Header(key, value)
        }
    }

    c.Status(resp.StatusCode)

    contentType := resp.Header.Get("Content-Type")
    if contentType == "application/json" {
        var jsonResponse interface{}
        if err := json.Unmarshal(responseBody, &jsonResponse); err == nil {
            c.JSON(resp.StatusCode, jsonResponse)
            return
        }
    }

    c.Data(resp.StatusCode, contentType, responseBody)
}