package routes

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/dragsbruh/hypersonic/internal/database"
)

func ParsePagination(r *http.Request, allowedColumns map[string]bool, defaultColumn string) (*database.PaginationConfig, error) {
	q := r.URL.Query()

	pageStr := q.Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		if pageStr == "" {
			page = 1
		} else {
			return nil, fmt.Errorf("bad page number")
		}
	}

	limitStr := q.Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		if limitStr == "" {
			limit = 50
		} else {
			return nil, fmt.Errorf("bad limit number")
		}
	}

	columnStr := q.Get("sort_by")
	if columnStr == "" {
		columnStr = defaultColumn
	} else if !allowedColumns[columnStr] {
		return nil, fmt.Errorf("column not allowed")
	}

	directionStr := q.Get("direction")

	if directionStr == "" {
		if columnStr == "random" {
			directionStr = strconv.Itoa(rand.Int())
		} else {
			directionStr = "asc"
		}
	}
	if columnStr != "random" {
		if !(directionStr == "asc" || directionStr == "desc") {
			return nil, fmt.Errorf("direction must be asc or desc")
		}
	}

	return &database.PaginationConfig{
		Column:    columnStr,
		Direction: directionStr,
		Page:      page,
		Limit:     limit,
	}, nil
}
