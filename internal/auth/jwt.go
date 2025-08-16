package auth

import (
	"fmt"
	"time"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

const UserID = "middleware.auth.UserID"
const CookieName = "hs-session"

func GetJwtUser(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return config.JWT_SECRET, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("unauthorized")
	}

	return claims.Subject, nil
}

func CreateJwtUser(userID string) (string, error) {
	expires := time.Now().Add(time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expires),
	})
	signedToken, err := token.SignedString(config.JWT_SECRET)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signedToken, nil
}
