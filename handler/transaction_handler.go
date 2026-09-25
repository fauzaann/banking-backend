package handler

import (
	"github.com/gin-gonic/gin"

	"banking/controller"
	"banking/middleware"
	"banking/pkg/response"
)

type TransactionHandler struct {
	ctrl controller.TransactionController
}

func NewTransactionHandler(ctrl controller.TransactionController) *TransactionHandler {
	return &TransactionHandler{ctrl: ctrl}
}

func buildTransactionQuery(c *gin.Context) controller.TransactionQuery {
	return controller.TransactionQuery{
		Type:      c.Query("type"),
		Status:    c.Query("status"),
		Reference: c.Query("reference"),
		StartDate: parseDateQuery(c, "start_date", false),
		EndDate:   parseDateQuery(c, "end_date", true),
	}
}

// List godoc: GET /api/v1/transactions
func (h *TransactionHandler) List(c *gin.Context) {
	p := parsePagination(c)
	res, err := h.ctrl.List(c.Request.Context(), middleware.GetUserID(c), buildTransactionQuery(c), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Data transaksi berhasil ditemukan", res, p)
}

// Detail godoc: GET /api/v1/transactions/:id
func (h *TransactionHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID transaksi tidak valid")
		return
	}

	res, err := h.ctrl.Detail(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Data transaksi berhasil ditemukan", res)
}

// AdminList godoc: GET /api/v1/admin/transactions
func (h *TransactionHandler) AdminList(c *gin.Context) {
	p := parsePagination(c)
	res, err := h.ctrl.ListAll(c.Request.Context(), buildTransactionQuery(c), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Data transaksi berhasil ditemukan", res, p)
}
