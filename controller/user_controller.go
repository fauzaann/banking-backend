package controller

import (
	"context"
	"time"

	"banking/models"
	apperr "banking/pkg/errors"
	"banking/pkg/password"
	"banking/repository"
)

type UserResponse struct {
	ID        uint      `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID: u.ID, FullName: u.FullName, Email: u.Email, Phone: u.Phone,
		Role: u.Role, Status: u.Status, CreatedAt: u.CreatedAt,
	}
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
	Phone    string `json:"phone" validate:"required,min=8,max=20"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UserController interface {
	GetProfile(ctx context.Context, userID uint) (*UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest, ip string) (*UserResponse, error)
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest, ip string) error
}

type userController struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
}

func NewUserController(userRepo repository.UserRepository, auditRepo repository.AuditRepository) UserController {
	return &userController{userRepo: userRepo, auditRepo: auditRepo}
}

func (c *userController) GetProfile(ctx context.Context, userID uint) (*UserResponse, error) {
	user, err := c.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}
	res := NewUserResponse(user)
	return &res, nil
}

func (c *userController) UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest, ip string) (*UserResponse, error) {
	user, err := c.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	user.FullName = req.FullName
	user.Phone = req.Phone
	if err := c.userRepo.Update(ctx, user); err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}

	recordAudit(ctx, c.auditRepo, uintPtr(userID), models.ActionUpdateProfile, "User memperbarui profil", ip)
	res := NewUserResponse(user)
	return &res, nil
}

func (c *userController) ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest, ip string) error {
	user, err := c.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if user == nil {
		return apperr.ErrUserNotFound
	}
	if !password.Compare(user.Password, req.OldPassword) {
		return apperr.ErrWrongPassword
	}

	hashed, err := password.Hash(req.NewPassword)
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	user.Password = hashed
	if err := c.userRepo.Update(ctx, user); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}

	recordAudit(ctx, c.auditRepo, uintPtr(userID), models.ActionChangePassword, "User mengganti password", ip)
	return nil
}
