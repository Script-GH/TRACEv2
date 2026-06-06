package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/trace/trace/internal/domain"
)

// JWTService issues and validates JWTs.
// 24-hour expiry. Re-issued on app open (spec requirement).
type JWTService struct {
	secret  []byte
	expiry  time.Duration
}

func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Issue creates a signed JWT for userID.
func (s *JWTService) Issue(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.expiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// Validate verifies the token signature and expiry, then returns the userID.
func (s *JWTService) Validate(tokenStr string) (userID string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return "", domain.ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", domain.ErrInvalidToken
	}
	id, ok := claims["user_id"].(string)
	if !ok || id == "" {
		return "", domain.ErrInvalidToken
	}
	return id, nil
}
