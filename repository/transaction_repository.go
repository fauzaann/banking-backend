package repository

import (
	"context"
	"time"
	"errors"

	"banking/pkg/utils"
	"banking/models"
	"gorm.io/gorm"
)

type TransactionFilter struct {
	AccountIDs []uint
	Type       string
	Status     string
	Reference  string
	StartDate  *time.Time
	EndDate    *time.Time
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *gorm.DB, trx *models.Transaction) error
	FindByID(ctx context.Context, id uint) (*models.Transaction, error)
	FindAll(ctx context.Context, filter TransactionFilter, p *utils.Pagination) ([]models.Transaction, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	SumAmountByStatus(ctx context.Context, status string) (int64, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) conn(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *transactionRepository) Create(ctx context.Context,tx *gorm.DB, trx *models.Transaction) error {
	return r.conn(tx).WithContext(ctx).Create(trx).Error
}

func (r *transactionRepository) FindByID(ctx context.Context, id uint) (*models.Transaction, error) {
	var tx models.Transaction
	if err := r.db.WithContext(ctx).First(&tx, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err	
	}
	return &tx, nil
}

func (r *transactionRepository) FindAll(ctx context.Context, filter TransactionFilter, p *utils.Pagination) ([]models.Transaction, error) {	
	query := r.db.WithContext(ctx).Model(&models.Transaction{})

	if filter.AccountIDs != nil {
		if len (filter.AccountIDs) == 0 {
			p.SetTotal(0)
			return []models.Transaction{}, nil
		}
		query = query.Where("account_id IN ?", filter.AccountIDs)
	}
	if filter.Type != "" {
		query = query.Where("transaction_type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Reference != "" {
		query = query.Where("reference_number ILIKE ?", "%"+filter.Reference+"%")
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	p.SetTotal(total)

	var trxs []models.Transaction
	err := query.Order("id DESC").Limit(p.Limit).Offset(p.Offset()).Find(&trxs).Error
	return trxs, err
}

func (r *transactionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Transaction{}).Count(&count).Error
	return count, err
}

func (r *transactionRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *transactionRepository) SumAmountByStatus(ctx context.Context, status string) (int64, error) {
	var result struct {
		Total int64
	}
	err := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) AS total").
		Where("status = ? AND transaction_type IN ?", status,
			[]string{models.TrxTypeTransferOut, models.TrxTypeWithdraw, models.TrxTypeDeposit}).
		Scan(&result).Error
	return result.Total, err
	
}