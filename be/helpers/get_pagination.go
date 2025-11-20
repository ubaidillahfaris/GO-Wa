package helpers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPagination(c *gin.Context, defaultLimit int64) (skip, limit int64) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	limit = defaultLimit
	page := int64(1)

	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 64); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	skip = (page - 1) * limit
	return
}
