package middleware

import (
	"context"
	"net/http"

	"github.com/dragsbruh/hypersonic/internal/auth"
)

func IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := auth.GetJwtUser(cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
