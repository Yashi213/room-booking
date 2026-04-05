package auth

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessTokenIssuer interface {
	SignAccessToken(userID uuid.UUID, role string) (string, error)
}

type HS256Access struct {
	secret []byte
	ttl    time.Duration
}

func NewHS256Access(secret []byte, ttl time.Duration) *HS256Access {
	return &HS256Access{secret: secret, ttl: ttl}
}

func (a *HS256Access) SignAccessToken(userID uuid.UUID, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(a.ttl).Unix(),
	})
	return token.SignedString(a.secret)
}

func (a *HS256Access) ParseBearerAuthorization(authHeader string) (role string, userID uuid.UUID, err error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", uuid.Nil, jwt.ErrTokenMalformed
	}
	rawToken := strings.TrimPrefix(authHeader, prefix)
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	if err != nil || !token.Valid {
		return "", uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", uuid.Nil, jwt.ErrTokenInvalidClaims
	}
	roleStr, ok := claims["role"].(string)
	if !ok || roleStr == "" {
		return "", uuid.Nil, jwt.ErrTokenInvalidClaims
	}
	userIDRaw, ok := claims["user_id"].(string)
	if !ok || userIDRaw == "" {
		return "", uuid.Nil, jwt.ErrTokenInvalidClaims
	}
	userID, err = uuid.Parse(userIDRaw)
	if err != nil {
		return "", uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	return roleStr, userID, nil
}
