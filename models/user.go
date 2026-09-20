package models

import (
	"time"
	
	"gorm.io/gorm"
)

const (
	RoleCustomer = "CUSTOMER"
	RoleAdmin    = "ADMIN"

	UserStatusActive  = "ACTIVE"
	UserStatusBlocked = "BLOCKED"
)

// User merepresentasikan entitas user dalam sistem.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	FullName  string         `gorm:"size:100;not null" json:"full_name"`
	Email     string         `gorm:"size:100;not null;uniqueIndex" json:"email"`
	Phone     string         `gorm:"size:20;not null" json:"phone"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      string         `gorm:"size:20;not null;default:CUSTOMER;index" json:"role"`
	Status    string         `gorm:"size:20;not null;default:ACTIVE;index" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Accounts      []Account     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"accounts,omitempty"`
	Beneficiaries []Beneficiary `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	AuditLogs     []AuditLog    `gorm:"foreignKey:UserID" json:"-"`
}

// TableName mengembalikan nama tabel untuk model User.
func (User) TableName() string { return "users" }

// IsAdmin memeriksa apakah user memiliki peran admin.
func (u *User) IsAdmin() bool  { return u.Role == RoleAdmin }
// IsActive memeriksa apakah user memiliki status aktif.
func (u *User) IsActive() bool { return u.Status == UserStatusActive }
