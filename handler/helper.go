package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"banking/pkg/utils"
)

// parsePagination membaca query ?page= &limit=
func parsePagination(c *gin.Context) *utils.Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	return utils.NewPagination(page, limit)
}

// parseIDParam membaca path param bertipe uint.
func parseIDParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// parseDateQuery menerima format YYYY-MM-DD.
func parseDateQuery(c *gin.Context, name string, endOfDay bool) *time.Time {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return &t
}
