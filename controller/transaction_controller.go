package controller

import (
	"context"
	"time"

	"banking/models"
	apperr "banking/pkg/errors"
	"banking/pkg/utils"
	"banking/repository"
)

// TransactionResponse adalah 
type TransactionResponse struct {
	ID              uint      `json:"id"`
	AccountID       uint      `json:"account_id"`
	TransactionCode string    `json:"transaction_code"`
	TransactionType string    `json:"transaction_type"`
	Amount          int64     `json:"amount"`
	BalanceBefore   int64     `json:"balance_before"`
	BalanceAfter    int64     `json:"balance_after"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	ReferenceNumber string    `json:"reference_number"`
	CreatedAt       time.Time `json:"created_at"`
}

func NewTransactionResponse(t *models.Transaction) TransactionResponse {
	return TransactionResponse{
		ID: t.ID, AccountID: t.AccountID, TransactionCode: t.TransactionCode,
		TransactionType: t.TransactionType, Amount: t.Amount,
		BalanceBefore: t.BalanceBefore, BalanceAfter: t.BalanceAfter,
		Description: t.Description, Status: t.Status,
		ReferenceNumber: t.ReferenceNumber, CreatedAt: t.CreatedAt,
	}
}

type TransactionQuery struct {
	Type      string
	Status    string
	Reference string
	StartDate *time.Time
	EndDate   *time.Time
}

type TransactionController interface {
	List(ctx context.Context, userID uint, q TransactionQuery, p *utils.Pagination) ([]TransactionResponse, error)
	Detail(ctx context.Context, userID, trxID uint) (*TransactionResponse, error)
	ListAll(ctx context.Context, q TransactionQuery, p *utils.Pagination) ([]TransactionResponse, error)
}

type transactionController struct {
	trxRepo     repository.TransactionRepository
	accountRepo repository.AccountRepository
}

func NewTransactionController(trxRepo repository.TransactionRepository, accountRepo repository.AccountRepository) TransactionController {
	return &transactionController{trxRepo: trxRepo, accountRepo: accountRepo}
}

func (c *transactionController) List(ctx context.Context, userID uint, q TransactionQuery, p *utils.Pagination) ([]TransactionResponse, error) {
	accounts, err := c.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	ids := make([]uint, 0, len(accounts))
	for i := range accounts {
		ids = append(ids, accounts[i].ID)
	}

	filter := repository.TransactionFilter{
		AccountIDs: ids, Type: q.Type, Status: q.Status,
		Reference: q.Reference, StartDate: q.StartDate, EndDate: q.EndDate,
	}
	trxs, err := c.trxRepo.FindAll(ctx, filter, p)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return mapTransactions(trxs), nil
}

func (c *transactionController) ListAll(ctx context.Context, q TransactionQuery, p *utils.Pagination) ([]TransactionResponse, error) {
	filter := repository.TransactionFilter{
		Type: q.Type, Status: q.Status, Reference: q.Reference,
		StartDate: q.StartDate, EndDate: q.EndDate,
	}
	trxs, err := c.trxRepo.FindAll(ctx, filter, p)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return mapTransactions(trxs), nil
}

func (c *transactionController) Detail(ctx context.Context, userID, trxID uint) (*TransactionResponse, error) {
	trx, err := c.trxRepo.FindByID(ctx, trxID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if trx == nil {
		return nil, apperr.ErrTransactionNotFound
	}

	account, err := c.accountRepo.FindByID(ctx, trx.AccountID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if account == nil || account.UserID != userID {
		return nil, apperr.ErrForbidden
	}

	res := NewTransactionResponse(trx)
	return &res, nil
}

func mapTransactions(trxs []models.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, 0, len(trxs))
	for i := range trxs {
		out = append(out, NewTransactionResponse(&trxs[i]))
	}
	return out
}
