package handler

import (
	"github.com/gin-gonic/gin"

	"banking/controller"
	"banking/middleware"
	"banking/pkg/response"
	"banking/pkg/validator"
)

type BeneficiaryHandler struct {
	ctrl      controller.BeneficiaryController
	validator *validator.Validator
}

func NewBeneficiaryHandler(ctrl controller.BeneficiaryController, v *validator.Validator) *BeneficiaryHandler {
	return &BeneficiaryHandler{ctrl: ctrl, validator: v}
}

// Create godoc: POST /api/v1/beneficiaries
func (h *BeneficiaryHandler) Create(c *gin.Context) {
	var req controller.CreateBeneficiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	res, err := h.ctrl.Create(c.Request.Context(), middleware.GetUserID(c), req, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Beneficiary berhasil ditambahkan", res)
}

// List godoc: GET /api/v1/beneficiaries
func (h *BeneficiaryHandler) List(c *gin.Context) {
	res, err := h.ctrl.List(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Data beneficiary berhasil ditemukan", res)
}

// Detail godoc: GET /api/v1/beneficiaries/:id
func (h *BeneficiaryHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID beneficiary tidak valid")
		return
	}

	res, err := h.ctrl.Detail(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Data beneficiary berhasil ditemukan", res)
}

// Delete godoc: DELETE /api/v1/beneficiaries/:id
func (h *BeneficiaryHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		response.BadRequest(c, "ID beneficiary tidak valid")
		return
	}

	if err := h.ctrl.Delete(c.Request.Context(), middleware.GetUserID(c), id, c.ClientIP()); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Beneficiary berhasil dihapus", nil)
}
