FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/api-gateway

# Use a smaller image for the final container
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the build stage
COPY --from=builder /app/api-gateway .
# COPY --from=builder /app/.env .  # Bu satırı kaldırın

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./api-gateway"]