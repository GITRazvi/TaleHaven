package jwt

import (
	"TaleHaven-Backend/modules/utils"
	"errors"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrMissingToken     = errors.New("token is missing")
	ErrInvalidSignature = errors.New("invalid token signature")
)

const (
	TokenDuration = 30 * 24 * time.Hour // 30 days
)

// Claims represents the JWT claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwtlib.RegisteredClaims
}

// getJWTSecret retrieves the JWT secret from environment variables
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		utils.Logger.Warn("JWT_SECRET not set, using default (NOT SECURE FOR PRODUCTION)")
		secret = "default-insecure-jwt-secret-change-this-in-production"
	}
	return []byte(secret)
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(TokenDuration)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			Issuer:    "talehaven-backend",
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		utils.Logger.Error("Failed to generate JWT token", zap.Error(err))
		return "", err
	}

	utils.Logger.Info("JWT token generated",
		zap.Int("user_id", userID),
		zap.String("email", email),
		zap.Time("expires_at", claims.ExpiresAt.Time))

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		utils.Logger.Warn("Failed to parse JWT token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func RefreshToken(oldTokenString string) (string, error) {
	claims, err := ValidateToken(oldTokenString)
	if err != nil {
		// Allow refresh of expired tokens (within reason)
		if !errors.Is(err, ErrExpiredToken) {
			return "", err
		}

		// Parse without validation to get claims
		token, _ := jwtlib.ParseWithClaims(oldTokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})

		if token == nil {
			return "", ErrInvalidToken
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			return "", ErrInvalidToken
		}

		// Only allow refresh if token expired less than 7 days ago
		if time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
			return "", errors.New("token expired too long ago")
		}
	}

	// Generate new token with same user info
	return GenerateToken(claims.UserID, claims.Email)
}

