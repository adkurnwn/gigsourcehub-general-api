package helpers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination holds the standard pagination variables parsed from a request
type Pagination struct {
	Page   int64
	Limit  int64
	Offset int64
	Cursor string
}

// GetPagination extracts page, limit, and cursor from the gin context safely
// Defaults to Page=1, Limit=10 if missing or invalid.
func GetPagination(c *gin.Context) Pagination {
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	cursor := c.Query("cursor")

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: offset,
		Cursor: cursor,
	}
}
