package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"banking/models"
	"banking/pkg/utils"
)

type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) error
	FindByID(ctx context.Context, id uint) (*models.Account, error)
	FindByNumber(ctx context.Context, number string) (*models.Account, error)
	FindByUserID(ctx context.Context, userID uint) ([]models.Account, error)
	FindAll(ctx context.Context, status string, p *utils.Pagination) ([]models.Account, error)
	ExistsByNumber(ctx context.Context, number string) (bool, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)

	// Dipakai di dalam database transaction (transfer).
	LockByID(ctx context.Context, tx *gorm.DB, id uint) (*models.Account, error)
	UpdateBalance(ctx context.Context, tx *gorm.DB, id uint, balance int64) error
}

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) AccountRepository { return &accountRepository{db: db} }

func (r *accountRepository) conn(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *accountRepository) Create(ctx context.Context, account *models.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *accountRepository) FindByID(ctx context.Context, id uint) (*models.Account, error) {
	var acc models.Account
	if err := r.db.WithContext(ctx).First(&acc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *accountRepository) FindByNumber(ctx context.Context, number string) (*models.Account, error) {
	var acc models.Account
	if err := r.db.WithContext(ctx).Where("account_number = ?", number).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *accountRepository) FindByUserID(ctx context.Context, userID uint) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) FindAll(ctx context.Context, status string, p *utils.Pagination) ([]models.Account, error) {
	query := r.db.WithContext(ctx).Model(&models.Account{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	p.SetTotal(total)

	var accounts []models.Account
	err := query.Order("id DESC").Limit(p.Limit).Offset(p.Offset()).Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) ExistsByNumber(ctx context.Context, number string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Account{}).
		Where("account_number = ?", number).Count(&total).Error
	return total > 0, err
}

func (r *accountRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Account{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *accountRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Account{}).Count(&total).Error
	return total, err
}

func (r *accountRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Account{}).Where("status = ?", status).Count(&total).Error
	return total, err
}

// LockByID mengunci baris rekening dengan SELECT ... FOR UPDATE.
func (r *accountRepository) LockByID(ctx context.Context, tx *gorm.DB, id uint) (*models.Account, error) {
	var acc models.Account
	err := r.conn(tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&acc, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &acc, nil
}

func (r *accountRepository) UpdateBalance(ctx context.Context, tx *gorm.DB, id uint, balance int64) error {
	return r.conn(tx).WithContext(ctx).Model(&models.Account{}).
		Where("id = ?", id).Update("balance", balance).Error
}
