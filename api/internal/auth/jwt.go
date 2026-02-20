// Package auth provides JWT token management and password hashing for Raisin Protect.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims holds the JWT payload with user identity and multi-tenant context.
type Claims struct {
	UserID string `json:"sub"`
	OrgID  string `json:"org"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// JWTManager handles JWT token operations.
type JWTManager struct {
	config JWTConfig
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(config JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

// TokenPair contains both access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateTokenPair creates both an access and refresh token.
func (m *JWTManager) GenerateTokenPair(userID, orgID, email, role string) (*TokenPair, error) {
	accessToken, err := m.generateToken(userID, orgID, email, role, TokenTypeAccess, m.config.AccessExpiry)
	if err != nil {
		return nil, err
	}
	refreshToken, err := m.generateToken(userID, orgID, email, role, TokenTypeRefresh, m.config.RefreshExpiry)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.config.AccessExpiry.Seconds()),
	}, nil
}

// GenerateAccessToken generates only an access token (for refresh flow).
func (m *JWTManager) GenerateAccessToken(userID, orgID, email, role string) (string, int64, error) {
	token, err := m.generateToken(userID, orgID, email, role, TokenTypeAccess, m.config.AccessExpiry)
	if err != nil {
		return "", 0, err
	}
	return token, int64(m.config.AccessExpiry.Seconds()), nil
}

func (m *JWTManager) generateToken(userID, orgID, email, role, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		OrgID:  orgID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ValidateToken parses and validates a JWT token.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.config.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ValidateAccessToken validates an access token specifically.
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeAccess {
		return nil, errors.New("not an access token")
	}
	return claims, nil
}

// ValidateRefreshToken validates a refresh token specifically.
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}
	return claims, nil
}
