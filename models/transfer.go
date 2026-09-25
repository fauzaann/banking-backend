package models

import "time"

const (
	TransferStatusPending = "PENDING"
	TransferStatusSuccess = "SUCCESS"
	TransferStatusFailed  = "FAILED"
)

type Transfer struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	SenderAccountID   uint      `gorm:"not null;index" json:"sender_account_id"`
	ReceiverAccountID uint      `gorm:"not null;index" json:"receiver_account_id"`
	Amount            int64     `gorm:"not null" json:"amount"`
	ReferenceNumber   string    `gorm:"size:30;not null;uniqueIndex" json:"reference_number"`
	Description       string    `gorm:"size:255" json:"description"`
	Status            string    `gorm:"size:20;not null;index" json:"status"`
	CreatedAt         time.Time `json:"created_at"`

	SenderAccount   *Account `gorm:"foreignKey:SenderAccountID" json:"sender_account,omitempty"`
	ReceiverAccount *Account `gorm:"foreignKey:ReceiverAccountID" json:"receiver_account,omitempty"`
}

func (Transfer) TableName() string { return "transfers" }
