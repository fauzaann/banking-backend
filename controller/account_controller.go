package controller

import (
	"context"

	"banking/models"
	apperr "banking/pkg/errors"
	"banking/repository"
)

// AccountResponse adalah representasi JSON dari model Account yang dikirimkan ke klien.
type AccountResponse struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	Balance       int64  `json:"balance"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}

// NewAccountResponse membuat instance AccountResponse dari model Account.
type BalanceResponse struct {
	AccountNumber string `json:"account_number"`
	Balance       int64  `json:"balance"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}

// NewAccountResponse membuat instance AccountResponse dari model Account.
func NewAccountResponse(a *models.Account) AccountResponse {
	return AccountResponse{
		ID: a.ID, UserID: a.UserID, AccountNumber: a.AccountNumber,
		AccountType: a.AccountType, Balance: a.Balance, Currency: a.Currency, Status: a.Status,
	}
}

// AccountController adalah antarmuka untuk mengelola operasi terkait rekening.
type AccountController interface {
	ListMyAccounts(ctx context.Context, userID uint) ([]AccountResponse, error)
	GetAccount(ctx context.Context, userID, accountID uint) (*AccountResponse, error)
	GetBalance(ctx context.Context, userID, accountID uint) (*BalanceResponse, error)
}

// accountController adalah implementasi dari AccountController.
type accountController struct {
	accountRepo repository.AccountRepository
}

// NewAccountController membuat instance baru dari accountController dengan repositori account yang diberikan.
func NewAccountController(accountRepo repository.AccountRepository) AccountController {
	return &accountController{accountRepo: accountRepo}
}

// ListMyAccounts mengambil daftar rekening milik user berdasarkan ID user yang diberikan.
func (c *accountController) ListMyAccounts(ctx context.Context, userID uint) ([]AccountResponse, error) {
	accounts, err := c.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]AccountResponse, 0, len(accounts))
	for i := range accounts {
		out = append(out, NewAccountResponse(&accounts[i]))
	}
	return out, nil
}

// findOwnedAccount memastikan rekening benar-benar milik user yang login.
func (c *accountController) findOwnedAccount(ctx context.Context, userID, accountID uint) (*models.Account, error) {
	account, err := c.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if account == nil {
		return nil, apperr.ErrAccountNotFound
	}
	if account.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	return account, nil
}

// GetAccount mengambil detail rekening milik user berdasarkan ID user dan ID rekening yang diberikan.
func (c *accountController) GetAccount(ctx context.Context, userID, accountID uint) (*AccountResponse, error) {
	account, err := c.findOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	res := NewAccountResponse(account)
	return &res, nil
}

// GetBalance mengambil saldo rekening milik user berdasarkan ID user dan ID rekening yang diberikan.
func (c *accountController) GetBalance(ctx context.Context, userID, accountID uint) (*BalanceResponse, error) {
	account, err := c.findOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	return &BalanceResponse{
		AccountNumber: account.AccountNumber,
		Balance:       account.Balance,
		Currency:      account.Currency,
		Status:        account.Status,
	}, nil
}
