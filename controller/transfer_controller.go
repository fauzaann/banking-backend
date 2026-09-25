package controller

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"banking/models"
	apperr "banking/pkg/errors"
	"banking/pkg/utils"
	"banking/repository"
)

type TransferRequest struct {
	SenderAccountID       uint   `json:"sender_account_id" validate:"required"`
	ReceiverAccountNumber string `json:"receiver_account_number" validate:"required,min=6,max=20"`
	Amount                int64  `json:"amount" validate:"required,gt=0"`
	Description           string `json:"description" validate:"max=255"`
}

type TransferResponse struct {
	TransactionID   uint      `json:"transaction_id"`
	TransferID      uint      `json:"transfer_id"`
	ReferenceNumber string    `json:"reference_number"`
	Amount          int64     `json:"amount"`
	SenderBalance   int64     `json:"sender_balance"`
	ReceiverAccount string    `json:"receiver_account"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type TransferListItem struct {
	ID                uint      `json:"id"`
	SenderAccountID   uint      `json:"sender_account_id"`
	ReceiverAccountID uint      `json:"receiver_account_id"`
	Amount            int64     `json:"amount"`
	ReferenceNumber   string    `json:"reference_number"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

func newTransferListItem(t *models.Transfer) TransferListItem {
	return TransferListItem{
		ID: t.ID, SenderAccountID: t.SenderAccountID, ReceiverAccountID: t.ReceiverAccountID,
		Amount: t.Amount, ReferenceNumber: t.ReferenceNumber, Description: t.Description,
		Status: t.Status, CreatedAt: t.CreatedAt,
	}
}

type TransferController interface {
	Transfer(ctx context.Context, userID uint, req TransferRequest, ip string) (*TransferResponse, error)
	List(ctx context.Context, userID uint, p *utils.Pagination) ([]TransferListItem, error)
	Detail(ctx context.Context, userID, transferID uint) (*TransferListItem, error)
}

type transferController struct {
	db          *gorm.DB
	accountRepo repository.AccountRepository
	transferRepo repository.TransferRepository
	trxRepo     repository.TransactionRepository
	auditRepo   repository.AuditRepository
}

func NewTransferController(
	db *gorm.DB,
	accountRepo repository.AccountRepository,
	transferRepo repository.TransferRepository,
	trxRepo repository.TransactionRepository,
	auditRepo repository.AuditRepository,
) TransferController {
	return &transferController{
		db: db, accountRepo: accountRepo, transferRepo: transferRepo,
		trxRepo: trxRepo, auditRepo: auditRepo,
	}
}

// ValidateTransfer memuat seluruh aturan bisnis transfer dalam bentuk fungsi
// murni sehingga mudah diuji tanpa database.
func ValidateTransfer(userID uint, sender, receiver *models.Account, amount int64) error {
	if amount <= 0 {
		return apperr.ErrInvalidAmount
	}
	if sender == nil {
		return apperr.ErrAccountNotFound
	}
	if sender.UserID != userID {
		return apperr.ErrForbidden
	}
	if receiver == nil {
		return apperr.ErrAccountNotFound
	}
	if sender.ID == receiver.ID {
		return apperr.ErrSelfTransfer
	}
	if err := assertUsable(sender); err != nil {
		return err
	}
	if err := assertUsable(receiver); err != nil {
		return err
	}
	if sender.Balance < amount {
		return apperr.ErrInsufficientBalance
	}
	return nil
}

func assertUsable(acc *models.Account) error {
	switch acc.Status {
	case models.AccountStatusFrozen:
		return apperr.ErrAccountFrozen
	case models.AccountStatusClosed:
		return apperr.ErrAccountClosed
	}
	return nil
}

// Transfer memindahkan dana antar rekening di dalam satu database transaction
// dengan row-level locking (SELECT ... FOR UPDATE) sehingga aman dari race
// condition, double spending, dan saldo negatif.
func (c *transferController) Transfer(ctx context.Context, userID uint, req TransferRequest, ip string) (*TransferResponse, error) {
	sender, err := c.accountRepo.FindByID(ctx, req.SenderAccountID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	receiver, err := c.accountRepo.FindByNumber(ctx, req.ReceiverAccountNumber)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	// Validasi awal (fail fast) sebelum membuka transaction.
	if err := ValidateTransfer(userID, sender, receiver, req.Amount); err != nil {
		return nil, err
	}

	var result TransferResponse

	txErr := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Kunci baris dengan urutan ID menaik untuk menghindari deadlock
		// ketika dua transfer berlawanan arah terjadi bersamaan.
		firstID, secondID := sender.ID, receiver.ID
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}
		lockedFirst, err := c.accountRepo.LockByID(ctx, tx, firstID)
		if err != nil {
			return err
		}
		lockedSecond, err := c.accountRepo.LockByID(ctx, tx, secondID)
		if err != nil {
			return err
		}
		if lockedFirst == nil || lockedSecond == nil {
			return apperr.ErrAccountNotFound
		}

		lockedSender, lockedReceiver := lockedFirst, lockedSecond
		if lockedSender.ID != sender.ID {
			lockedSender, lockedReceiver = lockedSecond, lockedFirst
		}

		// Validasi ulang setelah baris terkunci (kondisi bisa berubah).
		if err := ValidateTransfer(userID, lockedSender, lockedReceiver, req.Amount); err != nil {
			return err
		}

		senderBefore := lockedSender.Balance
		receiverBefore := lockedReceiver.Balance
		senderAfter := senderBefore - req.Amount
		receiverAfter := receiverBefore + req.Amount

		if err := c.accountRepo.UpdateBalance(ctx, tx, lockedSender.ID, senderAfter); err != nil {
			return err
		}
		if err := c.accountRepo.UpdateBalance(ctx, tx, lockedReceiver.ID, receiverAfter); err != nil {
			return err
		}

		reference := utils.GenerateReferenceNumber("TRX")
		transfer := &models.Transfer{
			SenderAccountID:   lockedSender.ID,
			ReceiverAccountID: lockedReceiver.ID,
			Amount:            req.Amount,
			ReferenceNumber:   reference,
			Description:       req.Description,
			Status:            models.TransferStatusSuccess,
		}
		if err := c.transferRepo.Create(ctx, tx, transfer); err != nil {
			return err
		}

		outTrx := &models.Transaction{
			AccountID:       lockedSender.ID,
			TransactionCode: utils.GenerateTransactionCode(),
			TransactionType: models.TrxTypeTransferOut,
			Amount:          req.Amount,
			BalanceBefore:   senderBefore,
			BalanceAfter:    senderAfter,
			Description:     fmt.Sprintf("Transfer ke %s", lockedReceiver.AccountNumber),
			Status:          models.TrxStatusSuccess,
			ReferenceNumber: reference,
		}
		if err := c.trxRepo.Create(ctx, tx, outTrx); err != nil {
			return err
		}

		inTrx := &models.Transaction{
			AccountID:       lockedReceiver.ID,
			TransactionCode: utils.GenerateTransactionCode(),
			TransactionType: models.TrxTypeTransferIn,
			Amount:          req.Amount,
			BalanceBefore:   receiverBefore,
			BalanceAfter:    receiverAfter,
			Description:     fmt.Sprintf("Transfer dari %s", lockedSender.AccountNumber),
			Status:          models.TrxStatusSuccess,
			ReferenceNumber: reference,
		}
		if err := c.trxRepo.Create(ctx, tx, inTrx); err != nil {
			return err
		}

		result = TransferResponse{
			TransactionID:   outTrx.ID,
			TransferID:      transfer.ID,
			ReferenceNumber: reference,
			Amount:          req.Amount,
			SenderBalance:   senderAfter,
			ReceiverAccount: lockedReceiver.AccountNumber,
			Status:          models.TransferStatusSuccess,
			CreatedAt:       transfer.CreatedAt,
		}
		return nil
	})

	// Jika ada error apa pun di dalam closure, GORM otomatis ROLLBACK.
	if txErr != nil {
		return nil, apperr.As(txErr)
	}

	recordAudit(ctx, c.auditRepo, uintPtr(userID), models.ActionTransfer,
		fmt.Sprintf("Transfer %d ke %s (%s)", req.Amount, result.ReceiverAccount, result.ReferenceNumber), ip)

	return &result, nil
}

func (c *transferController) List(ctx context.Context, userID uint, p *utils.Pagination) ([]TransferListItem, error) {
	accountIDs, err := c.userAccountIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	transfers, err := c.transferRepo.FindByAccountIDs(ctx, accountIDs, p)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]TransferListItem, 0, len(transfers))
	for i := range transfers {
		out = append(out, newTransferListItem(&transfers[i]))
	}
	return out, nil
}

func (c *transferController) Detail(ctx context.Context, userID, transferID uint) (*TransferListItem, error) {
	transfer, err := c.transferRepo.FindByID(ctx, transferID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if transfer == nil {
		return nil, apperr.ErrTransferNotFound
	}

	accountIDs, err := c.userAccountIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	owned := false
	for _, id := range accountIDs {
		if id == transfer.SenderAccountID || id == transfer.ReceiverAccountID {
			owned = true
			break
		}
	}
	if !owned {
		return nil, apperr.ErrForbidden
	}

	item := newTransferListItem(transfer)
	return &item, nil
}

func (c *transferController) userAccountIDs(ctx context.Context, userID uint) ([]uint, error) {
	accounts, err := c.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	ids := make([]uint, 0, len(accounts))
	for i := range accounts {
		ids = append(ids, accounts[i].ID)
	}
	return ids, nil
}
