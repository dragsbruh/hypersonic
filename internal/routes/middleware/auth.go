package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/dragsbruh/hypersonic/internal/auth"
	"github.com/dragsbruh/hypersonic/internal/config"
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

func MustAdmin(next http.Handler) http.Handler {
	return MustAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.AuthUserID).(string)
		if !slices.Contains(config.Administrators, userID) {
			http.Error(w, "unauthorized (not admin)", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}))
}
