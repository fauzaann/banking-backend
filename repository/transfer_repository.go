package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"banking/models"
	"banking/pkg/utils"
)

type TransferRepository interface {
	Create(ctx context.Context, tx *gorm.DB, transfer *models.Transfer) error
	FindByID(ctx context.Context, id uint) (*models.Transfer, error)
	FindByAccountIDs(ctx context.Context, accountIDs []uint, p *utils.Pagination) ([]models.Transfer, error)
}

type transferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) TransferRepository { return &transferRepository{db: db} }

func (r *transferRepository) conn(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *transferRepository) Create(ctx context.Context, tx *gorm.DB, transfer *models.Transfer) error {
	return r.conn(tx).WithContext(ctx).Create(transfer).Error
}

func (r *transferRepository) FindByID(ctx context.Context, id uint) (*models.Transfer, error) {
	var transfer models.Transfer
	err := r.db.WithContext(ctx).
		Preload("SenderAccount").Preload("ReceiverAccount").
		First(&transfer, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *transferRepository) FindByAccountIDs(ctx context.Context, accountIDs []uint, p *utils.Pagination) ([]models.Transfer, error) {
	if len(accountIDs) == 0 {
		p.SetTotal(0)
		return []models.Transfer{}, nil
	}

	query := r.db.WithContext(ctx).Model(&models.Transfer{}).
		Where("sender_account_id IN ? OR receiver_account_id IN ?", accountIDs, accountIDs)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	p.SetTotal(total)

	var transfers []models.Transfer
	err := query.Order("id DESC").Limit(p.Limit).Offset(p.Offset()).Find(&transfers).Error
	return transfers, err
}
