package router

import (
	"net/http"

	"github.com/dragsbruh/hypersonic/internal/router/api"
)

func Router() http.Handler {
	r := http.NewServeMux()

	r.Handle("/api", api.Router())

	return r
}
