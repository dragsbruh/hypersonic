package api

import (
	"context"
	"net/http"

	"github.com/dragsbruh/hypersonic/internal/database"
)

func Router(ctx context.Context, queries *database.PGQueries) http.Handler {
	_ = &Handler{
		MainCtx: ctx,
		Queries: queries,
	}

	r := http.NewServeMux()

	return r
}
