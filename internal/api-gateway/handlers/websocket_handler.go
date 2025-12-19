package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"whatsapp/pkg/auth"
	"whatsapp/pkg/models"
	"whatsapp/pkg/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles websocket connections
type WebSocketHandler struct {
    messageServiceURL string
    upgrader         websocket.Upgrader
    clients          map[string]*websocket.Conn
    clientsMutex     sync.RWMutex
    rabbitMQClient   *rabbitmq.Client
    authService      *auth.Service
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(messageServiceURL string, rabbitMQClient *rabbitmq.Client, authService *auth.Service) *WebSocketHandler {
    handler := &WebSocketHandler{
        messageServiceURL: messageServiceURL,
        clients:          make(map[string]*websocket.Conn),
        clientsMutex:     sync.RWMutex{},
        rabbitMQClient:   rabbitMQClient,
        authService:      authService,
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin: func(r *http.Request) bool {
                return true
            },
        },
    }
    
    // Set up RabbitMQ consumer for WebSocket delivery
    if rabbitMQClient != nil {
        // Declare exchange for messages
        if err := rabbitMQClient.DeclareExchange("messages", "topic"); err != nil {
            log.Printf("Failed to declare exchange: %v", err)
        }
        
        // Declare queue for WebSocket messages
        queue, err := rabbitMQClient.DeclareQueue("websocket_messages")
        if err != nil {
            log.Printf("Failed to declare queue: %v", err)
        }
        
        // Bind queue to exchange with routing patterns
        if err = rabbitMQClient.BindQueue(queue.Name, "message.#", "messages"); err != nil {
            log.Printf("Failed to bind queue: %v", err)
        }

        if err = rabbitMQClient.BindQueue(queue.Name, "status.#", "messages"); err != nil {
            log.Printf("Failed to bind queue: %v", err)
        }

        // Bind typing events
        if err = rabbitMQClient.BindQueue(queue.Name, "typing.#", "messages"); err != nil {
            log.Printf("Failed to bind typing queue: %v", err)
        }

        log.Printf("WebSocket Handler: RabbitMQ Consumer Setup Complete")
        
        // Start consuming messages
        if err = rabbitMQClient.Consume(queue.Name, handler.handleIncomingRabbitMQMessage); err != nil {
            log.Printf("Failed to start consuming messages: %v", err)
        }
    } else {
        log.Printf("CRITICAL: WebSocketHandler initialized with NIL RabbitMQ client. Real-time messaging will NOT work.")
    }
    
    return handler
}

// HandleWebSocket handles websocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    token := c.Query("token")
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
        return
    }

    claims, err := h.authService.ValidateToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
        return
    }

    UserID := claims.UserID
    c.Set("UserID", UserID)
    UserIDStr := UserID

    log.Printf("WebSocket connection attempt from user: %s", UserIDStr)

    h.clientsMutex.Lock()
    existingConn, exists := h.clients[UserIDStr]
    if exists {
        log.Printf("Closing existing connection for user: %s", UserIDStr)
        existingConn.Close()
        delete(h.clients, UserIDStr)
    }
    h.clientsMutex.Unlock()

    conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }

    log.Printf("WebSocket connection established for user: %s", UserIDStr)

    conn.SetPingHandler(func(pingData string) error {
        log.Printf("Received ping from user %s", UserIDStr)
        return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(10*time.Second))
    })
    
    conn.SetPongHandler(func(pongData string) error {
        log.Printf("Received pong from user %s", UserIDStr)
        return nil
    })

    h.clientsMutex.Lock()
    h.clients[UserIDStr] = conn
    h.clientsMutex.Unlock()

    if h.rabbitMQClient != nil {
        statusUpdate := models.StatusUpdate{Status: "online"}
        routingKey := fmt.Sprintf("status.user.%s", UserIDStr)
        if err := h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate); err != nil {
            log.Printf("Failed to publish online status: %v", err)
        }
    }

    pingTicker := time.NewTicker(30 * time.Second)

    defer func() {
        pingTicker.Stop()
        conn.Close()
        h.clientsMutex.Lock()
        delete(h.clients, UserIDStr)
        h.clientsMutex.Unlock()
        
        log.Printf("WebSocket connection closed for user: %s", UserIDStr)

        if h.rabbitMQClient != nil {
            statusUpdate := models.StatusUpdate{Status: "offline"}
            routingKey := fmt.Sprintf("status.user.%s", UserIDStr)
            if err := h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate); err != nil {
                log.Printf("Failed to publish offline status: %v", err)
            }
        }
    }()

    go func() {
        for range pingTicker.C {
            if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
                log.Printf("Error sending ping to user %s: %v", UserIDStr, err)
                return
            }
            log.Printf("Ping sent to user: %s", UserIDStr)
        }
    }()

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Error reading message from user %s: %v", UserIDStr, err)
            break
        }
        
        if messageType == websocket.TextMessage && string(p) == "ping" {
            if err := conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
                log.Printf("Error sending pong to user %s: %v", UserIDStr, err)
                break
            }
            continue
        }

        if messageType == websocket.TextMessage {
            var baseMsg map[string]interface{}
            if err := json.Unmarshal(p, &baseMsg); err != nil {
                log.Printf("Error unmarshalling message: %v", err)
                continue
            }

            if msgType, ok := baseMsg["type"].(string); ok && msgType == "typing" {
                var typingEvent models.TypingEvent
                if err := json.Unmarshal(p, &typingEvent); err != nil {
                    log.Printf("Error unmarshalling typing event: %v", err)
                    continue
                }

                typingEvent.SenderID = UserIDStr
                typingEvent.Timestamp = time.Now().Format(time.RFC3339)

                if h.rabbitMQClient != nil {
                    routingKey := fmt.Sprintf("typing.%s", typingEvent.ReceiverID)
                    if err := h.rabbitMQClient.PublishToExchange("messages", routingKey, typingEvent); err != nil {
                        log.Printf("Failed to publish typing event: %v", err)
                    }
                } else {
                    h.sendTypingEventDirect(typingEvent)
                }
                continue
            }

            var msg models.MessageRequest
            if err := json.Unmarshal(p, &msg); err != nil {
                log.Printf("Error unmarshalling message: %v", err)
                continue
            }

            authHeader := c.Request.Header.Get("Authorization")
            if authHeader == "" {
                authHeader = "Bearer " + token
            }
            
            h.sendMessageViaHTTP(msg, authHeader)
        }
    }
}

// sendTypingEventDirect sends typing event directly to WebSocket client
func (h *WebSocketHandler) sendTypingEventDirect(event models.TypingEvent) {
    h.clientsMutex.RLock()
    defer h.clientsMutex.RUnlock()

    if conn, ok := h.clients[event.ReceiverID]; ok {
        if err := conn.WriteJSON(event); err != nil {
            log.Printf("Error sending typing event to WebSocket: %v", err)
        }
    }
}

// sendMessageViaHTTP sends a message payload using HTTP to the message service
func (h *WebSocketHandler) sendMessageViaHTTP(payload interface{}, authHeader string) {
    reqBody, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshalling message payload: %v", err)
        return
    }

    req, err := http.NewRequest("POST", h.messageServiceURL+"/messages", bytes.NewBuffer(reqBody))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        return
    }

    req.Header.Set("Content-Type", "application/json")
    if authHeader != "" {
        req.Header.Set("Authorization", authHeader)
    }

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending message to message service: %v", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        log.Printf("Message service returned error: %d - %s", resp.StatusCode, string(body))
    }
}

// handleIncomingRabbitMQMessage processes messages from RabbitMQ and forwards them to WebSocket clients
func (h *WebSocketHandler) handleIncomingRabbitMQMessage(body []byte) error {
    var msg map[string]interface{}
    if err := json.Unmarshal(body, &msg); err != nil {
        log.Printf("Error unmarshalling RabbitMQ message: %v", err)
        return err
    }

    if msgType, ok := msg["type"].(string); ok && msgType == "typing" {
        if receiverID, ok := msg["receiver_id"].(string); ok {
            h.clientsMutex.RLock()
            if conn, ok := h.clients[receiverID]; ok {
                if err := conn.WriteJSON(msg); err != nil {
                    log.Printf("Error sending typing event to WebSocket: %v", err)
                }
            }
            h.clientsMutex.RUnlock()
        }
        return nil
    }

    if msgType, ok := msg["type"].(string); ok && msgType == "batch" {
        if senderID, ok := msg["sender_id"].(string); ok {
            h.clientsMutex.RLock()
            if conn, ok := h.clients[senderID]; ok {
                if err := conn.WriteJSON(msg); err != nil {
                    log.Printf("Error sending batch update to WebSocket: %v", err)
                }
            }
            h.clientsMutex.RUnlock()
        }
        return nil
    }

    if _, ok := msg["content"].(string); ok {
        if receiverID, ok := msg["receiver_id"].(string); ok {
            h.clientsMutex.RLock()
            if conn, ok := h.clients[receiverID]; ok {
                if err := conn.WriteJSON(msg); err != nil {
                    log.Printf("Error sending message to WebSocket: %v", err)
                }
            }
            h.clientsMutex.RUnlock()
        }
        return nil
    }

    if _, ok := msg["message_id"].(string); ok {
        if senderID, ok := msg["sender_id"].(string); ok {
            h.clientsMutex.RLock()
            if conn, ok := h.clients[senderID]; ok {
                if err := conn.WriteJSON(msg); err != nil {
                    log.Printf("Error sending status update to WebSocket: %v", err)
                }
            }
            h.clientsMutex.RUnlock()
        }
        return nil
    }

    if _, ok := msg["status"].(string); ok {
         if _, ok := msg["UserID"].(string); ok {
         }
         return nil
    }

    return nil
}

// SendMessage forwards a message to the message service via HTTP or RabbitMQ
func (h *WebSocketHandler) SendMessage(c *gin.Context) {
    UserID, exists := c.Get("UserID")
    if (!exists) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var msgReq models.MessageRequest
    if err := c.ShouldBindJSON(&msgReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    message := struct {
        SenderID   string `json:"sender_id"`
        ReceiverID string `json:"receiver_id"`
        Content    string `json:"content"`
        MediaURL   string `json:"media_url,omitempty"`
    }{
        SenderID:   UserID.(string),
        ReceiverID: msgReq.ReceiverID,
        Content:    msgReq.Content,
        MediaURL:   msgReq.MediaURL,
    }

    // Try to send via RabbitMQ if available
    if h.rabbitMQClient != nil {
        routingKey := fmt.Sprintf("message.%s", msgReq.ReceiverID)
        err := h.rabbitMQClient.PublishToExchange("messages", routingKey, message)
        if err == nil {
            c.JSON(http.StatusCreated, gin.H{"status": "Message sent"})
            return
        }
        log.Printf("Failed to publish to RabbitMQ: %v - falling back to HTTP", err)
    }

    // Fallback to HTTP
    reqBody, err := json.Marshal(message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
        return
    }

    resp, err := http.Post(h.messageServiceURL+"/messages", "application/json", bytes.NewBuffer(reqBody))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Message service unavailable"})
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
        return
    }

    c.Data(resp.StatusCode, "application/json", body)
}

// GetMessageHistory forwards the message history request to the message service
func (h *WebSocketHandler) GetMessageHistory(c *gin.Context) {
    UserID, exists := c.Get("UserID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    otherUserID := c.Param("UserID")
    
    url := fmt.Sprintf("%s/messages/%s?with=%s", h.messageServiceURL, UserID.(string), otherUserID)
    
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", c.GetHeader("Authorization"))
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Message service unavailable"})
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
        return
    }

    c.Data(resp.StatusCode, "application/json", body)
}
