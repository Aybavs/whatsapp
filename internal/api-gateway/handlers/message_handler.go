package handlers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageHandler handles message-related requests in the API gateway
type MessageHandler struct {
    messageServiceURL string
}

// NewMessageHandler creates a new message handler for the API gateway
func NewMessageHandler(messageServiceURL string) *MessageHandler {
    return &MessageHandler{
        messageServiceURL: messageServiceURL,
    }
}

// SendMessage forwards message creation requests to the message service
func (h *MessageHandler) SendMessage(c *gin.Context) {
    h.proxyRequest(c, "/messages", http.MethodPost)
}

// GetMessages retrieves messages for a specific user conversation
func (h *MessageHandler) GetMessages(c *gin.Context) {
    UserID := c.Param("UserID")
    h.proxyRequest(c, "/messages/"+UserID+"?"+c.Request.URL.RawQuery, http.MethodGet)
}

// UpdateMessageStatus handles message status updates (read, delivered)
func (h *MessageHandler) UpdateMessageStatus(c *gin.Context) {
    messageID := c.Param("id")
    h.proxyRequest(c, "/messages/"+messageID+"/status", http.MethodPatch)
}

// SearchMessages forwards search requests to the message service
func (h *MessageHandler) SearchMessages(c *gin.Context) {
    h.proxyRequest(c, "/messages/search?"+c.Request.URL.RawQuery, http.MethodGet)
}

// proxyRequest forwards the request to the message service
func (h *MessageHandler) proxyRequest(c *gin.Context, path string, method string) {
    var requestBody []byte
    if c.Request.Body != nil {
        requestBody, _ = io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
    }

    req, err := http.NewRequest(method, h.messageServiceURL+path, bytes.NewBuffer(requestBody))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
        return
    }

    req.Header = c.Request.Header
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Message service unavailable"})
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