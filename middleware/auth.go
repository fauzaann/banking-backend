package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"banking/models"
	apperr "banking/pkg/errors"
	jwtpkg "banking/pkg/jwt"
	"banking/pkg/response"
	"banking/repository"
)

const (
	CtxUserID    = "user_id"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
	CtxClaims    = "jwt_claims"
)

// Auth memvalidasi JWT, memeriksa blacklist, dan memastikan user masih aktif.
func Auth(manager *jwtpkg.Manager, userRepo repository.UserRepository, blacklist *jwtpkg.Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}

		claims, err := manager.Parse(parts[1])
		if err != nil || claims.Type != jwtpkg.TokenTypeAccess {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}
		if blacklist.IsRevoked(claims.ID) {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			response.Error(c, apperr.ErrInternal.Wrap(err))
			return
		}
		if user == nil {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}
		if user.Status != models.UserStatusActive {
			response.Error(c, apperr.ErrUserBlocked)
			return
		}

		c.Set(CtxUserID, user.ID)
		c.Set(CtxUserEmail, user.Email)
		c.Set(CtxUserRole, user.Role)
		c.Set(CtxClaims, claims)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func GetUserRole(c *gin.Context) string {
	if v, ok := c.Get(CtxUserRole); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

func GetClaims(c *gin.Context) *jwtpkg.Claims {
	if v, ok := c.Get(CtxClaims); ok {
		if claims, ok := v.(*jwtpkg.Claims); ok {
			return claims
		}
	}
	return nil
}
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"banking-backend/models"
	apperr "banking-backend/pkg/errors"
	jwtpkg "banking-backend/pkg/jwt"
	"banking-backend/pkg/response"
	"banking-backend/repository"
)

const (
	CtxUserID    = "user_id"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
	CtxClaims    = "jwt_claims"
)

// Auth memvalidasi JWT, memeriksa blacklist, dan memastikan user masih aktif.
func Auth(manager *jwtpkg.Manager, userRepo repository.UserRepository, blacklist *jwtpkg.Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}

		claims, err := manager.Parse(parts[1])
		if err != nil || claims.Type != jwtpkg.TokenTypeAccess {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}
		if blacklist.IsRevoked(claims.ID) {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			response.Error(c, apperr.ErrInternal.Wrap(err))
			return
		}
		if user == nil {
			response.Error(c, apperr.ErrUnauthorized)
			return
		}
		if user.Status != models.UserStatusActive {
			response.Error(c, apperr.ErrUserBlocked)
			return
		}

		c.Set(CtxUserID, user.ID)
		c.Set(CtxUserEmail, user.Email)
		c.Set(CtxUserRole, user.Role)
		c.Set(CtxClaims, claims)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func GetUserRole(c *gin.Context) string {
	if v, ok := c.Get(CtxUserRole); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

func GetClaims(c *gin.Context) *jwtpkg.Claims {
	if v, ok := c.Get(CtxClaims); ok {
		if claims, ok := v.(*jwtpkg.Claims); ok {
			return claims
		}
	}
	return nil
}
