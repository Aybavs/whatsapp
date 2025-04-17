# WhatsApp Clone Backend - Go Microservices Architecture

A microservices-based backend for a chat application built with Go.

## Architecture

This project uses a microservices architecture with:

- **API Gateway**: Routes HTTP and WebSocket requests to appropriate microservices
- **User Service**: Handles user registration, authentication, and user management
- **Message Service**: Handles message exchange and storage
- **RabbitMQ**: Used for asynchronous communication between services
- **MongoDB**: Used for data storage

## Running the Application

### Prerequisites

- Docker and Docker Compose
- Go 1.21 or higher (for local development)

### Using Docker Compose

```bash
docker-compose up -d
```

This will start all the necessary services:

- MongoDB on port 27017
- RabbitMQ on ports 5672 (AMQP) and 15672 (Management UI)
- API Gateway on port 8080
- User Service on port 8081
- Message Service on port 8082

## API Endpoints

### Public Endpoints

- `POST /api/register`: Register a new user
- `POST /api/login`: Login and receive a JWT token

### Protected Endpoints (Require JWT Authentication)

- `GET /api/users`: Search for users
- `GET /api/users/:id`: Get user details
- `GET /api/ws`: WebSocket endpoint for real-time messaging
- `GET /api/messages/:UserID`: Get message history with another user
- `POST /api/messages`: Send a message via REST API

## Authentication

The application uses JWT tokens for authentication. After logging in, include the token in subsequent requests as follows:

```
Authorization: Bearer <your-token>
```

## WebSocket Communication

After authenticating, connect to the WebSocket endpoint at `/api/ws` with the JWT token in the request header. WebSocket messages should be JSON objects with the following format:

```json
{
  "receiver_id": "user-id-to-send-message-to",
  "content": "Message content"
}
```
