package api

import (
	"context"

	"github.com/dragsbruh/hypersonic/internal/database"
)

type Handler struct {
	Queries *database.PGQueries
	MainCtx context.Context
}
