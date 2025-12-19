package handlers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GroupHandler handles group-related requests in the API gateway
type GroupHandler struct {
    userServiceURL string
}

// NewGroupHandler creates a new group handler for the API gateway
func NewGroupHandler(userServiceURL string) *GroupHandler {
    return &GroupHandler{
        userServiceURL: userServiceURL,
    }
}

// CreateGroup proxies a request to create a new group
func (h *GroupHandler) CreateGroup(c *gin.Context) {
    h.proxyRequest(c, "/groups", http.MethodPost)
}

// GetUserGroups proxies a request to get groups for the user
func (h *GroupHandler) GetUserGroups(c *gin.Context) {
    h.proxyRequest(c, "/groups", http.MethodGet)
}

// proxyRequest forwards the request to the user service
// Duplicated from UserHandler for simplicity to avoid circular deps or common pkg overhead for now
func (h *GroupHandler) proxyRequest(c *gin.Context, path string, method string) {
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
