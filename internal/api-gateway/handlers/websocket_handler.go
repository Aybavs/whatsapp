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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebSocketHandler handles websocket connections
type WebSocketHandler struct {
    messageServiceURL string
    upgrader         websocket.Upgrader
    clients          map[string]*websocket.Conn
    clientsMutex     sync.RWMutex
    rabbitMQClient   *rabbitmq.Client
    authService      *auth.Service  // Auth service eklendi
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(messageServiceURL string, rabbitMQClient *rabbitmq.Client, authService *auth.Service) *WebSocketHandler {
    handler := &WebSocketHandler{
        messageServiceURL: messageServiceURL,
        clients:          make(map[string]*websocket.Conn),
        clientsMutex:     sync.RWMutex{},
        rabbitMQClient:   rabbitMQClient,
        authService:      authService,  // Auth service yapılandırması
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin: func(r *http.Request) bool {
                // Geliştirme ortamında tüm origin'lere izin ver
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
        
        // Start consuming messages
        if err = rabbitMQClient.Consume(queue.Name, handler.handleIncomingRabbitMQMessage); err != nil {
            log.Printf("Failed to start consuming messages: %v", err)
        }
    }
    
    return handler
}

// HandleWebSocket handles websocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    // Token parametresinden kullanıcıyı doğrula
    token := c.Query("token")
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
        return
    }
    
    // Token'ı doğrula ve kullanıcı ID'sini al
    claims, err := h.authService.ValidateToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
        return
    }
    
    // Kullanıcı ID'sini gin context'e ekle
    UserID := claims.UserID
    c.Set("UserID", UserID)
    UserIDStr := UserID
    
    log.Printf("WebSocket connection attempt from user: %s", UserIDStr)
    
    // Mevcut bağlantıyı kontrol et ve kapat
    h.clientsMutex.Lock()
    existingConn, exists := h.clients[UserIDStr]
    if exists {
        log.Printf("Closing existing connection for user: %s", UserIDStr)
        existingConn.Close()
        delete(h.clients, UserIDStr)
    }
    h.clientsMutex.Unlock()
    
    // Geriye kalan WebSocket işlemleri...
    conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }

    log.Printf("WebSocket connection established for user: %s", UserIDStr)
    
    // Bağlantı kurulduktan sonra, ping-pong yapılandırması
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
    
    // Publish online status via RabbitMQ
    if h.rabbitMQClient != nil {
        statusUpdate := models.StatusUpdate{Status: "online"}
        routingKey := fmt.Sprintf("status.user.%s", UserIDStr)
        if err := h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate); err != nil {
            log.Printf("Failed to publish online status: %v", err)
        }
    }
    
    // Ping-pong mekanizması için ticker oluştur (30 saniyede bir ping gönder)
    pingTicker := time.NewTicker(30 * time.Second)
    
    // Bağlantı kapandığında temizleme işlemi
    defer func() {
        pingTicker.Stop()
        conn.Close()
        h.clientsMutex.Lock()
        delete(h.clients, UserIDStr)
        h.clientsMutex.Unlock()
        
        log.Printf("WebSocket connection closed for user: %s", UserIDStr)
        
        // Publish offline status via RabbitMQ
        if h.rabbitMQClient != nil {
            statusUpdate := models.StatusUpdate{Status: "offline"}
            routingKey := fmt.Sprintf("status.user.%s", UserIDStr)
            if err := h.rabbitMQClient.PublishToExchange("messages", routingKey, statusUpdate); err != nil {
                log.Printf("Failed to publish offline status: %v", err)
            }
        }
    }()
    
    // Ping-pong için ayrı bir goroutine başlat
    go func() {
        for range pingTicker.C {
            // Ping gönder
            if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
                log.Printf("Error sending ping to user %s: %v", UserIDStr, err)
                return
            }
            log.Printf("Ping sent to user: %s", UserIDStr)
        }
    }()

    // WebSocket mesajlarını işleme
    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Error reading message from user %s: %v", UserIDStr, err)
            break
        }
        
        // Ping-Pong mesajlarını yönet
        if messageType == websocket.TextMessage && string(p) == "ping" {
            if err := conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
                log.Printf("Error sending pong to user %s: %v", UserIDStr, err)
                break
            }
            continue
        }

        if messageType == websocket.TextMessage {
            var msg models.MessageRequest
            if err := json.Unmarshal(p, &msg); err != nil {
                log.Printf("Error unmarshalling message: %v", err)
                continue
            }

            senderObjectID, err := primitive.ObjectIDFromHex(UserIDStr)
            if err != nil {
                log.Printf("Invalid sender ID: %v", err)
                continue
            }

            receiverObjectID, err := primitive.ObjectIDFromHex(msg.ReceiverID)
            if err != nil {
                log.Printf("Invalid receiver ID: %v", err)
                continue
            }

            message := models.Message{
                SenderID:   senderObjectID,
                ReceiverID: receiverObjectID,
                Content:    msg.Content,
                MediaURL:   msg.MediaURL,
            }

            // Send message through RabbitMQ instead of direct HTTP if available
            if h.rabbitMQClient != nil {
                routingKey := fmt.Sprintf("message.%s", msg.ReceiverID)
                if err := h.rabbitMQClient.PublishToExchange("messages", routingKey, message); err != nil {
                    log.Printf("Failed to publish message to RabbitMQ: %v", err)
                    
                    // Fallback to HTTP if RabbitMQ fails
                    h.sendMessageViaHTTP(message)
                }
            } else {
                // Use HTTP if RabbitMQ client is not available
                h.sendMessageViaHTTP(message)
            }
        }
    }
}

// sendMessageViaHTTP sends a message using HTTP to the message service
func (h *WebSocketHandler) sendMessageViaHTTP(message models.Message) {
    reqBody, err := json.Marshal(message)
    if err != nil {
        log.Printf("Error marshalling message: %v", err)
        return
    }

    resp, err := http.Post(h.messageServiceURL+"/messages", "application/json", bytes.NewBuffer(reqBody))
    if err != nil {
        log.Printf("Error sending message to message service: %v", err)
        return
    }
    defer resp.Body.Close()
}

// handleIncomingRabbitMQMessage processes messages from RabbitMQ and forwards them to WebSocket clients
func (h *WebSocketHandler) handleIncomingRabbitMQMessage(body []byte) error {
    var msg map[string]interface{}
    if err := json.Unmarshal(body, &msg); err != nil {
        return err
    }
    
    // Determine message type and extract recipient ID
    if receiverID, ok := msg["receiver_id"].(string); ok {
        // This is a regular message
        h.clientsMutex.RLock()
        if conn, ok := h.clients[receiverID]; ok {
            if err := conn.WriteJSON(msg); err != nil {
                log.Printf("Error sending message to WebSocket: %v", err)
            }
        }
        h.clientsMutex.RUnlock()
    } else if messageID, ok := msg["message_id"].(string); ok {
        // This is a message status update
        // Extract sender ID from message ID - would need a DB lookup in reality
        // For now, we just log it
        log.Printf("Status update for message %s: %v", messageID, msg["status"])
    } else if statusUpdate, ok := msg["status"].(string); ok {
        // This might be a user status update
        if UserID, ok := msg["UserID"].(string); ok {
            log.Printf("User %s status changed to %s", UserID, statusUpdate)
        }
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