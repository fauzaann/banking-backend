package models

import "time"

const (
	AccountTypeSavings = "SAVINGS"
	AccountTypeCurrent = "CURRENT"

	AccountStatusActive = "ACTIVE"
	AccountStatusFrozen = "FROZEN"
	AccountStatusClosed = "CLOSED"
)

// Balance disimpan sebagai int64 (satuan rupiah penuh, tanpa pecahan desimal)
// supaya tidak terjadi pembulatan seperti pada float64.
type Account struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	AccountNumber string    `gorm:"size:20;not null;uniqueIndex" json:"account_number"`
	AccountType   string    `gorm:"size:20;not null;default:SAVINGS" json:"account_type"`
	Balance       int64     `gorm:"not null;default:0;check:balance >= 0" json:"balance"`
	Currency      string    `gorm:"size:3;not null;default:IDR" json:"currency"`
	Status        string    `gorm:"size:20;not null;default:ACTIVE;index" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:AccountID" json:"-"`
}

func (Account) TableName() string { return "accounts" }

func (a *Account) IsActive() bool { return a.Status == AccountStatusActive }
