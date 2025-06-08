// internal/utils/jwt.go
package utils

import (
	"errors"
	"time"

	"eticketing/internal/config"
	"eticketing/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	UserType models.UserType `json:"user_type"`
	Type     string          `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type JWTManager struct {
	config *config.JWTConfig
}

func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{config: cfg}
}

func (j *JWTManager) GenerateAccessToken(userID uint, username, email string, userType models.UserType) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		UserType: userType,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.AccessDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

func (j *JWTManager) GenerateRefreshToken(userID uint, username, email string, userType models.UserType) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		UserType: userType,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.RefreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	if claims.Type != "refresh" {
		return "", errors.New("invalid token type")
	}

	// Create new access token with updated expiration
	newClaims := JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		UserType: claims.UserType,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   claims.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.AccessDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString([]byte(j.config.Secret))
}
