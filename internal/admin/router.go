package admin

import (
	"context"
	"net/http"

	"github.com/dragsbruh/hypersonic/internal/database"
)

func Router(ctx context.Context, queries *database.PGQueries) http.Handler {
	handler := &Handler{
		MainCtx: ctx,
		Queries: queries,
	}

	r := http.NewServeMux()

	r.HandleFunc("GET /library/scan", handler.Scan)

	return r
}
