package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"banking/models"
	"banking/pkg/utils"
)

// UserFilter digunakan untuk memfilter hasil query user berdasarkan role, status, dan pencarian nama/email.
type UserFilter struct {
	Role   string
	Status string
	Search string
}

// UserRepository mendefinisikan kontrak untuk operasi database terkait entitas User.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	FindAll(ctx context.Context, filter UserFilter, p *utils.Pagination) ([]models.User, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// userRepository adalah implementasi UserRepository menggunakan GORM.
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID mencari user berdasarkan ID. Jika tidak ditemukan, mengembalikan nil.
func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail mencari user berdasarkan email. Jika tidak ditemukan, mengembalikan nil.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update memperbarui data user yang sudah ada di database.
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateStatus memperbarui status user berdasarkan ID.
func (r *userRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).
		Update("status", status).Error
}

// FindAll mengambil daftar user berdasarkan filter dan pagination
func (r *userRepository) FindAll(ctx context.Context, filter UserFilter, p *utils.Pagination) ([]models.User, error) {
	query := r.db.WithContext(ctx).Model(&models.User{})
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("full_name ILIKE ? OR email ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	p.SetTotal(total)

	var users []models.User
	err := query.Order("id DESC").Limit(p.Limit).Offset(p.Offset()).Find(&users).Error
	return users, err
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error
	return total, err
}

func (r *userRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("status = ?", status).Count(&total).Error
	return total, err
}
