package repository

import (
	"context"
	"errors"

	"banking/models"
	"gorm.io/gorm"
)

type BeneficiaryRepository interface {
	Create(ctx context.Context, beneficiary *models.Beneficiary) error
	FindByID(ctx context.Context, id uint) (*models.Beneficiary, error)
	FindByUserIDAndAccountNumber(ctx context.Context, userID uint, accountNumber string) (*models.Beneficiary, error)
	FindAllByUserID(ctx context.Context, userID uint) ([]models.Beneficiary, error)
	Delete(ctx context.Context, beneficiary *models.Beneficiary) error
}

type beneficiaryRepository struct {
	db *gorm.DB
}

func NewBeneficiaryRepository(db *gorm.DB) BeneficiaryRepository {
	return &beneficiaryRepository{db: db}
}

func (r *beneficiaryRepository) Create(ctx context.Context, b *models.Beneficiary) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *beneficiaryRepository) FindByID(ctx context.Context, id uint) (*models.Beneficiary, error) {
	var b models.Beneficiary
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *beneficiaryRepository) FindByUserID(ctx context.Context, userID uint) ([]models.Beneficiary, error) {
	var list []models.Beneficiary
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *beneficiaryRepository) ExistsByUserAndNumber(ctx context.Context, userID uint, number string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Beneficiary{}).
		Where("user_id = ? AND account_number = ?", userID, number).Count(&total).Error
	return total > 0, err
}

func (r *beneficiaryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Beneficiary{}, id).Error
}
