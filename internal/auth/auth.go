package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

const CookieName = "hs-session"
const AuthUserID = "hs.userID"

func CreateJWTUser(userID string, expiresAfter time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresAfter)),
	})
	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

func GetJWTUser(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return config.JWTSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.Subject, nil
}
