package api

import (
	"net/http"
)

func Router() http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("POST /auth/login", loginRoute)

	return r
}
