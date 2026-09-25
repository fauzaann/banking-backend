package handler

import (
	"github.com/gin-gonic/gin"

	"banking/controller"
	"banking/middleware"
	"banking/pkg/response"
	"banking/pkg/validator"
)

type TransferHandler struct {
	ctrl      controller.TransferController
	validator *validator.Validator
}

func NewTransferHandler(ctrl controller.TransferController, v *validator.Validator) *TransferHandler {
	return &TransferHandler{ctrl: ctrl, validator: v}
}

// Create godoc: POST /api/v1/transfers
func (h *TransferHandler) Create(c *gin.Context) {
	var req controller.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	res, err := h.ctrl.Transfer(c.Request.Context(), middleware.GetUserID(c), req, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Transfer berhasil", res)
}

// List godoc: GET /api/v1/transfers
func (h *TransferHandler) List(c *gin.Context) {
	p := parsePagination(c)
	res, err := h.ctrl.List(c.Request.Context(), middleware.GetUserID(c), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Data transfer berhasil ditemukan", res, p)
}

// Detail godoc: GET /api/v1/transfers/:id
func (h *TransferHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID transfer tidak valid")
		return
	}

	res, err := h.ctrl.Detail(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Data transfer berhasil ditemukan", res)
}
