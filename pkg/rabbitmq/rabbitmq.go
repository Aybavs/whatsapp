package rabbitmq

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
    conn         *amqp.Connection
    channel      *amqp.Channel
    uri          string
    clientMutex  sync.RWMutex
    queues       map[string]amqp.Queue      // Track declared queues
    exchanges    map[string]string          // Track declared exchanges by name->type
    bindings     map[string][]bindingInfo   // Track queue bindings
}

type bindingInfo struct {
    QueueName    string
    RoutingKey   string
    ExchangeName string
}

// NewClient creates a new RabbitMQ client
func NewClient(uri string) (*Client, error) {
    conn, err := amqp.Dial(uri)
    if err != nil {
        return nil, err
    }

    channel, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    client := &Client{
        conn:       conn, 
        channel:    channel,
        uri:        uri,
        queues:     make(map[string]amqp.Queue),
        exchanges:  make(map[string]string),
        bindings:   make(map[string][]bindingInfo),
    }
    
    // Set up reconnection handling
    go client.handleReconnect()
    
    return client, nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
    if err := c.channel.Close(); err != nil {
        return err
    }
    return c.conn.Close()
}

// DeclareQueue declares a new queue
func (c *Client) DeclareQueue(name string) (amqp.Queue, error) {
    queue, err := c.channel.QueueDeclare(
        name,  // name
        true,  // durable
        false, // delete when unused
        false, // exclusive
        false, // no-wait
        nil,   // arguments
    )
    
    if err == nil {
        c.clientMutex.Lock()
        c.queues[name] = queue
        c.clientMutex.Unlock()
    }
    
    return queue, err
}

// DeclareQueueWithDLX declares a queue with a dead-letter exchange
func (c *Client) DeclareQueueWithDLX(name, dlxName string) (amqp.Queue, error) {
    args := amqp.Table{
        "x-dead-letter-exchange": dlxName,
    }
    
    queue, err := c.channel.QueueDeclare(
        name,  // name
        true,  // durable
        false, // delete when unused
        false, // exclusive
        false, // no-wait
        args,  // arguments with dead letter exchange
    )
    
    if err == nil {
        c.clientMutex.Lock()
        c.queues[name] = queue
        c.clientMutex.Unlock()
    }
    
    return queue, err
}

// DeclareExchange declares a new exchange
func (c *Client) DeclareExchange(name string, exchangeType string) error {
    err := c.channel.ExchangeDeclare(
        name,         // name
        exchangeType, // type (direct, fanout, topic, headers)
        true,         // durable
        false,        // auto-deleted
        false,        // internal
        false,        // no-wait
        nil,          // arguments
    )
    
    if err == nil {
        c.clientMutex.Lock()
        c.exchanges[name] = exchangeType
        c.clientMutex.Unlock()
    }
    
    return err
}

// BindQueue binds a queue to an exchange with a routing key
func (c *Client) BindQueue(queueName, routingKey, exchangeName string) error {
    err := c.channel.QueueBind(
        queueName,    // queue name
        routingKey,   // routing key
        exchangeName, // exchange
        false,        // no-wait
        nil,          // arguments
    )
    
    if err == nil {
        c.clientMutex.Lock()
        binding := bindingInfo{
            QueueName:    queueName,
            RoutingKey:   routingKey,
            ExchangeName: exchangeName,
        }
        c.bindings[queueName] = append(c.bindings[queueName], binding)
        c.clientMutex.Unlock()
    }
    
    return err
}

// Publish publishes a message to a queue directly
func (c *Client) Publish(queue string, data interface{}) error {
    body, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return c.channel.Publish(
        "",    // exchange
        queue, // routing key
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
        },
    )
}

// PublishToExchange publishes a message to an exchange with a routing key
func (c *Client) PublishToExchange(exchange, routingKey string, data interface{}) error {
    body, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return c.channel.Publish(
        exchange,   // exchange
        routingKey, // routing key
        false,      // mandatory
        false,      // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
        },
    )
}

// Consume consumes messages from the specified queue
func (c *Client) Consume(queue string, handler func([]byte) error) error {
    msgs, err := c.channel.Consume(
        queue, // queue
        "",    // consumer
        false, // auto-ack
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        return err
    }

    go func() {
        for msg := range msgs {
            err := handler(msg.Body)
            if err != nil {
                log.Printf("Error processing message: %v", err)
                msg.Nack(false, true) // Nack the message and requeue
            } else {
                msg.Ack(false) // Ack the message
            }
        }
    }()

    return nil
}

// DeleteQueue deletes a queue if it exists
func (c *Client) DeleteQueue(name string) error {
    _, err := c.channel.QueueDelete(
        name,  // queue name
        false, // ifUnused (false = delete even if in use)
        false, // ifEmpty (false = delete even if not empty)
        false, // noWait
    )
    
    if err != nil {
        return err
    }
    
    c.clientMutex.Lock()
    delete(c.queues, name)
    c.clientMutex.Unlock()
    
    return nil
}

// DeleteExchange deletes an exchange if it exists
func (c *Client) DeleteExchange(name string) error {
    return c.channel.ExchangeDelete(
        name,  // exchange name
        false, // ifUnused (false = delete even if in use)
        false, // noWait
    )
}

// handleReconnect monitors the connection and reconnects if it drops
func (c *Client) handleReconnect() {
    // Set up notification channel for connection close
    connCloseChan := make(chan *amqp.Error)
    c.conn.NotifyClose(connCloseChan)
    
    // Wait for connection close event
    <-connCloseChan
    log.Println("RabbitMQ connection closed. Attempting to reconnect...")
    
    for {
        // Try to reconnect
        conn, err := amqp.Dial(c.uri)
        if err != nil {
            log.Printf("Failed to reconnect to RabbitMQ: %v. Retrying in 5 seconds...", err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        channel, err := conn.Channel()
        if err != nil {
            log.Printf("Failed to create channel: %v. Retrying in 5 seconds...", err)
            conn.Close()
            time.Sleep(5 * time.Second)
            continue
        }
        
        // Update connection and channel
        c.clientMutex.Lock()
        oldConn := c.conn
        c.conn = conn
        c.channel = channel
        c.clientMutex.Unlock()
        
        // Close old connection
        if oldConn != nil {
            _ = oldConn.Close()
        }
        
        log.Println("Successfully reconnected to RabbitMQ")
        
        // Redeclare exchanges
        c.clientMutex.RLock()
        for name, exchangeType := range c.exchanges {
            err = c.channel.ExchangeDeclare(
                name, exchangeType, true, false, false, false, nil)
            if err != nil {
                log.Printf("Failed to redeclare exchange %s: %v", name, err)
            }
        }
        
        // Redeclare queues
        for name := range c.queues {
            _, err = c.channel.QueueDeclare(
                name, true, false, false, false, nil)
            if err != nil {
                log.Printf("Failed to redeclare queue %s: %v", name, err)
            }
        }
        
        // Rebind queues
        for _, bindings := range c.bindings {
            for _, binding := range bindings {
                err = c.channel.QueueBind(
                    binding.QueueName, binding.RoutingKey, binding.ExchangeName, false, nil)
                if err != nil {
                    log.Printf("Failed to rebind queue %s to exchange %s: %v", 
                        binding.QueueName, binding.ExchangeName, err)
                }
            }
        }
        c.clientMutex.RUnlock()
        
        // Set up notification for the new connection
        connCloseChan = make(chan *amqp.Error)
        c.conn.NotifyClose(connCloseChan)
        
        // Wait for new connection close event
        <-connCloseChan
        log.Println("RabbitMQ connection closed. Attempting to reconnect...")
    }
}