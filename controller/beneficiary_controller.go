package controller

import (
	"context"
	"time"

	"banking/models"
	apperr "banking/pkg/errors"
	"banking/repository"
)

type CreateBeneficiaryRequest struct {
	AccountNumber string `json:"account_number" validate:"required,min=6,max=20"`
	Nickname      string `json:"nickname" validate:"required,min=2,max=100"`
}

type BeneficiaryResponse struct {
	ID            uint      `json:"id"`
	AccountNumber string    `json:"account_number"`
	Nickname      string    `json:"nickname"`
	OwnerName     string    `json:"owner_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type BeneficiaryController interface {
	Create(ctx context.Context, userID uint, req CreateBeneficiaryRequest, ip string) (*BeneficiaryResponse, error)
	List(ctx context.Context, userID uint) ([]BeneficiaryResponse, error)
	Detail(ctx context.Context, userID, id uint) (*BeneficiaryResponse, error)
	Delete(ctx context.Context, userID, id uint, ip string) error
}

type beneficiaryController struct {
	benefRepo   repository.BeneficiaryRepository
	accountRepo repository.AccountRepository
	userRepo    repository.UserRepository
	auditRepo   repository.AuditRepository
}

func NewBeneficiaryController(
	benefRepo repository.BeneficiaryRepository,
	accountRepo repository.AccountRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
) BeneficiaryController {
	return &beneficiaryController{
		benefRepo: benefRepo, accountRepo: accountRepo,
		userRepo: userRepo, auditRepo: auditRepo,
	}
}

func (c *beneficiaryController) Create(ctx context.Context, userID uint, req CreateBeneficiaryRequest, ip string) (*BeneficiaryResponse, error) {
	// Rekening tujuan wajib benar-benar ada dan tidak tertutup.
	account, err := c.accountRepo.FindByNumber(ctx, req.AccountNumber)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if account == nil {
		return nil, apperr.ErrAccountNotFound
	}
	if account.Status == models.AccountStatusClosed {
		return nil, apperr.ErrAccountClosed
	}
	if account.UserID == userID {
		return nil, apperr.ErrSelfTransfer
	}

	exists, err := c.benefRepo.ExistsByUserAndNumber(ctx, userID, req.AccountNumber)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if exists {
		return nil, apperr.ErrDuplicateBenef
	}

	benef := &models.Beneficiary{
		UserID:        userID,
		AccountNumber: req.AccountNumber,
		Nickname:      req.Nickname,
	}
	if err := c.benefRepo.Create(ctx, benef); err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}

	recordAudit(ctx, c.auditRepo, uintPtr(userID), models.ActionAddBeneficiary,
		"Menambah beneficiary "+req.AccountNumber, ip)

	res := c.toResponse(ctx, benef)
	return &res, nil
}

func (c *beneficiaryController) List(ctx context.Context, userID uint) ([]BeneficiaryResponse, error) {
	list, err := c.benefRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]BeneficiaryResponse, 0, len(list))
	for i := range list {
		out = append(out, c.toResponse(ctx, &list[i]))
	}
	return out, nil
}

func (c *beneficiaryController) Detail(ctx context.Context, userID, id uint) (*BeneficiaryResponse, error) {
	benef, err := c.find(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	res := c.toResponse(ctx, benef)
	return &res, nil
}

func (c *beneficiaryController) Delete(ctx context.Context, userID, id uint, ip string) error {
	benef, err := c.find(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := c.benefRepo.Delete(ctx, benef.ID); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	recordAudit(ctx, c.auditRepo, uintPtr(userID), models.ActionDelBeneficiary,
		"Menghapus beneficiary "+benef.AccountNumber, ip)
	return nil
}

func (c *beneficiaryController) find(ctx context.Context, userID, id uint) (*models.Beneficiary, error) {
	benef, err := c.benefRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if benef == nil {
		return nil, apperr.ErrBeneficiaryNotFound
	}
	if benef.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	return benef, nil
}

func (c *beneficiaryController) toResponse(ctx context.Context, b *models.Beneficiary) BeneficiaryResponse {
	res := BeneficiaryResponse{
		ID: b.ID, AccountNumber: b.AccountNumber,
		Nickname: b.Nickname, CreatedAt: b.CreatedAt,
	}
	if account, err := c.accountRepo.FindByNumber(ctx, b.AccountNumber); err == nil && account != nil {
		if owner, err := c.userRepo.FindByID(ctx, account.UserID); err == nil && owner != nil {
			res.OwnerName = owner.FullName
		}
	}
	return res
}
