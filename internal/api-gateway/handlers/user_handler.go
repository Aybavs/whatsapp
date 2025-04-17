package handlers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related requests in the API gateway
type UserHandler struct {
    userServiceURL string
}

// NewUserHandler creates a new user handler for the API gateway
func NewUserHandler(userServiceURL string) *UserHandler {
    return &UserHandler{
        userServiceURL: userServiceURL,
    }
}

// GetUserByID proxies a request to get a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
    UserID := c.Param("id")
    h.proxyRequest(c, "/users/"+UserID, http.MethodGet)
}

// SearchUsers proxies a request to search for users
func (h *UserHandler) SearchUsers(c *gin.Context) {
    h.proxyRequest(c, "/users/search?"+c.Request.URL.RawQuery, http.MethodGet)
}

// UpdateProfile proxies a request to update a user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
    UserID := c.Param("id")
    h.proxyRequest(c, "/users/"+UserID, http.MethodPut)
}

// UpdateStatus proxies a request to update a user's status
func (h *UserHandler) UpdateStatus(c *gin.Context) {
    UserID := c.Param("id")
    h.proxyRequest(c, "/users/"+UserID+"/status", http.MethodPatch)
}

// GetUserContacts proxies a request to get contacts (users with chat history)
func (h *UserHandler) GetUserContacts(c *gin.Context) {
    h.proxyRequest(c, "/users/contacts", http.MethodGet)
}

// AddContact proxies a request to add a contact
func (h *UserHandler) AddContact(c *gin.Context) {
    h.proxyRequest(c, "/users/contacts", http.MethodPost)
}

// DeleteContact proxies a request to delete a contact
func (h *UserHandler) DeleteContact(c *gin.Context) {
    contactID := c.Param("id")
    h.proxyRequest(c, "/users/contacts/"+contactID, http.MethodDelete)
}

// proxyRequest forwards the request to the user service
func (h *UserHandler) proxyRequest(c *gin.Context, path string, method string) {
    var requestBody []byte
    if c.Request.Body != nil {
        requestBody, _ = io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
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

    for key, values := range resp.Header {
        for _, value := range values {
            c.Header(key, value)
        }
    }

    c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)
}