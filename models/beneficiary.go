package models

import "time"

type Beneficiary struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index:idx_beneficiary_user_account,unique" json:"user_id"`
	AccountNumber string    `gorm:"size:20;not null;index:idx_beneficiary_user_account,unique" json:"account_number"`
	Nickname      string    `gorm:"size:100;not null" json:"nickname"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (Beneficiary) TableName() string { return "beneficiaries" }
