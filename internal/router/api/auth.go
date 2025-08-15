package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dragsbruh/hypersonic/internal/auth"
	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *loginRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)

	if r.Username == "" || r.Password == "" {
		return fmt.Errorf("missing username/password")
	}

	// TODO: validate username/password

	return nil
}

func loginRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data loginRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "error decoding json", http.StatusBadRequest)
		return
	}

	if err := data.Validate(); err != nil {
		// i like titlecase json sincee its written in go it feels neat
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"Status":  "Error",
			"Message": err.Error(),
		})
		return
	}

	user, err := auth.CheckUserLogin(ctx, data.Username, data.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		logrus.Errorf("error checking login: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	expires := time.Now().Add(time.Hour * 24 * 7)

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expires.Unix(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.JWTSecret)
	if err != nil {
		logrus.Errorf("error generating jwt secret: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "hs_jwt",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
		// no secure as intended to be used with reverse proxy as tls terminator
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func authMiddleware(required bool, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("hs_jwt")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return 
		}

		jwt.
	})
}
