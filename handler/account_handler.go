package handler

import (
	"github.com/gin-gonic/gin"

	"banking/controller"
	"banking/middleware"
	"banking/models"
	"banking/pkg/response"
	"banking/repository"
)

type AdminHandler struct {
	ctrl controller.AdminController
}

func NewAdminHandler(ctrl controller.AdminController) *AdminHandler {
	return &AdminHandler{ctrl: ctrl}
}

// Dashboard godoc: GET /api/v1/admin/dashboard
func (h *AdminHandler) Dashboard(c *gin.Context) {
	res, err := h.ctrl.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Statistik banking berhasil ditemukan", res)
}

// ListUsers godoc: GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	p := parsePagination(c)
	filter := repository.UserFilter{
		Role:   c.Query("role"),
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	res, err := h.ctrl.ListUsers(c.Request.Context(), filter, p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Data user berhasil ditemukan", res, p)
}

// GetUser godoc: GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID user tidak valid")
		return
	}

	res, err := h.ctrl.GetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Data user berhasil ditemukan", res)
}

// BlockUser godoc: PUT /api/v1/admin/users/:id/block
func (h *AdminHandler) BlockUser(c *gin.Context) {
	h.setUserStatus(c, models.UserStatusBlocked, "User berhasil diblokir")
}

// ActivateUser godoc: PUT /api/v1/admin/users/:id/activate
func (h *AdminHandler) ActivateUser(c *gin.Context) {
	h.setUserStatus(c, models.UserStatusActive, "User berhasil diaktifkan")
}

func (h *AdminHandler) setUserStatus(c *gin.Context, status, message string) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID user tidak valid")
		return
	}

	err := h.ctrl.SetUserStatus(c.Request.Context(), middleware.GetUserID(c), id, status, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, message, nil)
}

// ListAccounts godoc: GET /api/v1/admin/accounts
func (h *AdminHandler) ListAccounts(c *gin.Context) {
	p := parsePagination(c)
	res, err := h.ctrl.ListAccounts(c.Request.Context(), c.Query("status"), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Data rekening berhasil ditemukan", res, p)
}

// GetAccount godoc: GET /api/v1/admin/accounts/:id
func (h *AdminHandler) GetAccount(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID rekening tidak valid")
		return
	}

	res, err := h.ctrl.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Data rekening berhasil ditemukan", res)
}

// FreezeAccount godoc: PUT /api/v1/admin/accounts/:id/freeze
func (h *AdminHandler) FreezeAccount(c *gin.Context) {
	h.setAccountStatus(c, models.AccountStatusFrozen, "Rekening berhasil dibekukan")
}

// ActivateAccount godoc: PUT /api/v1/admin/accounts/:id/activate
func (h *AdminHandler) ActivateAccount(c *gin.Context) {
	h.setAccountStatus(c, models.AccountStatusActive, "Rekening berhasil diaktifkan")
}

func (h *AdminHandler) setAccountStatus(c *gin.Context, status, message string) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID rekening tidak valid")
		return
	}

	err := h.ctrl.SetAccountStatus(c.Request.Context(), middleware.GetUserID(c), id, status, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, message, nil)
}

// ListAuditLogs godoc: GET /api/v1/admin/audit-logs
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	p := parsePagination(c)
	res, err := h.ctrl.ListAuditLogs(c.Request.Context(), c.Query("action"), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Audit log berhasil ditemukan", res, p)
}
