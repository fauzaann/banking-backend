package models

import "time"

const (
	TrxTypeDeposit     = "DEPOSIT"
	TrxTypeWithdraw    = "WITHDRAW"
	TrxTypeTransferIn  = "TRANSFER_IN"
	TrxTypeTransferOut = "TRANSFER_OUT"

	TrxStatusPending = "PENDING"
	TrxStatusSuccess = "SUCCESS"
	TrxStatusFailed  = "FAILED"
)

type Transaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AccountID       uint      `gorm:"not null;index" json:"account_id"`
	TransactionCode string    `gorm:"size:30;not null;uniqueIndex" json:"transaction_code"`
	TransactionType string    `gorm:"size:20;not null;index" json:"transaction_type"`
	Amount          int64     `gorm:"not null" json:"amount"`
	BalanceBefore   int64     `gorm:"not null" json:"balance_before"`
	BalanceAfter    int64     `gorm:"not null" json:"balance_after"`
	Description     string    `gorm:"size:255" json:"description"`
	Status          string    `gorm:"size:20;not null;index" json:"status"`
	ReferenceNumber string    `gorm:"size:30;index" json:"reference_number"`
	CreatedAt       time.Time `json:"created_at"`

	Account *Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (Transaction) TableName() string { return "transactions" }
