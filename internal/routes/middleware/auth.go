package middleware

import (
	"context"
	"net/http"

	"github.com/dragsbruh/hypersonic/internal/auth"
	"github.com/sirupsen/logrus"
)

func MustAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie(auth.CookieName)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := auth.GetJWTUser(token.Value)
		if err != nil {
			logrus.Errorf("error getting jwt user: %v", err)
			http.Error(w, "jwt error", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.AuthUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
