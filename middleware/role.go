package middleware

import (
	"github.com/gin-gonic/gin"

	apperr "banking/pkg/errors"
	"banking/pkg/response"
)

// RequireRole membatasi endpoint hanya untuk role tertentu, misal RequireRole("ADMIN").
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role := GetUserRole(c)
		if _, ok := allowed[role]; !ok {
			response.Error(c, apperr.ErrForbidden)
			return
		}
		c.Next()
	}
}
