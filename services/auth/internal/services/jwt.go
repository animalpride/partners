package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/config"
	"github.com/dgrijalva/jwt-go"
)

// JWTService provides methods for generating and validating JWT tokens
type JWTService struct {
	cfg *config.Config
}

// NewJWTService creates a new JWTService
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		cfg: cfg,
	}
}

// GenerateAccessToken generates a short-lived JWT access token for a user
func (s *JWTService) GenerateAccessToken(userID int, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"typ":     "access",
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	return s.signClaims(claims)
}

// GenerateRefreshToken generates a long-lived refresh token for a user
func (s *JWTService) GenerateRefreshToken(userID int, familyID, tokenID string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"typ":       "refresh",
		"family_id": familyID,
		"jti":       tokenID,
		"exp":       time.Now().Add(ttl).Unix(),
		"iat":       time.Now().Unix(),
	}
	return s.signClaims(claims)
}

// ValidateAccessToken validates an access token and returns the user ID.
func (s *JWTService) ValidateAccessToken(tokenString string) (int, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return 0, err
	}
	if claims["typ"] != "access" {
		return 0, errors.New("invalid token type")
	}
	return userIDFromClaims(claims)
}

// ValidateRefreshToken validates a refresh token and returns its claims.
func (s *JWTService) ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims["typ"] != "refresh" {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

func (s *JWTService) signClaims(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (s *JWTService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func userIDFromClaims(claims jwt.MapClaims) (int, error) {
	// Handle user_id which comes as float64 from JSON
	if userIDFloat, ok := claims["user_id"].(float64); ok {
		return int(userIDFloat), nil
	}
	if userIDInt, ok := claims["user_id"].(int); ok {
		return userIDInt, nil
	}
	if userIDStr, ok := claims["user_id"].(string); ok {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return 0, err
		}
		return userID, nil
	}
	return 0, errors.New("invalid user_id format in token")
}
