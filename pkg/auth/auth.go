package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service provides JWT token generation and validation
type Service struct {
	secretKey     string
	tokenDuration time.Duration
}

// TokenClaims represents the claims in the JWT token
type TokenClaims struct {
	UserID   string `json:"UserID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewService creates a new authentication service
func NewService(secretKey string, tokenDuration time.Duration) *Service {
	return &Service{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a new JWT token for the given user
func (s *Service) GenerateToken(UserID, username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(s.tokenDuration)

	claims := &TokenClaims{
		UserID:   UserID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))

	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ValidateToken validates a JWT token and returns the claims if valid
func (s *Service) ValidateToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetUserIDFromToken extracts the user ID from a token string
func (s *Service) GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// GetUsernameFromToken extracts the username from a token string
func (s *Service) GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}
