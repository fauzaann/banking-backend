package controller

import (
	"context"
	"time"

	"banking/models"
	apperr "banking/pkg/errors"
	jwtpkg "banking/pkg/jwt"
	"banking/pkg/password"
	"banking/pkg/utils"
	"banking/repository"
)

// RegisterRequest adalah struktur data yang digunakan untuk permintaan registrasi user baru.
type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,min=8,max=20"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest adalah struktur data yang digunakan untuk permintaan login user.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshReqquest adalah struktur daya yang digunakan untuk permintaan refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse adalah struktur data yang digunakan untuk merespons permintaan autentikasi.
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// AuthController adalah antarmuka untuk mengelola operasi autentikasi user.
type AuthController interface {
	Register(ctx context.Context, req RegisterRequest, ip string) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest, ip string) (*AuthResponse, error)
	Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error)
	Logout(ctx context.Context, claims *jwtpkg.Claims, ip string) error
}

// authController adalah implementasi dari AuthController.
type authController struct {
	userRepo    repository.UserRepository
	accountRepo repository.AccountRepository
	auditRepo   repository.AuditRepository
	jwtManager  *jwtpkg.Manager
	blacklist   *jwtpkg.Blacklist
}

// NewAuthController membuat instance baru dari authController dengan repositori user, account, audit, jwtManager, dan blacklist yang diberikan.
func NewAuthController(
	userRepo repository.UserRepository,
	accountRepo repository.AccountRepository,
	auditRepo repository.AuditRepository,
	jwtManager *jwtpkg.Manager,
	blacklist *jwtpkg.Blacklist,
) AuthController {
	return &authController{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		auditRepo:   auditRepo,
		jwtManager:  jwtManager,
		blacklist:   blacklist,
	}
}

// Register membuat user CUSTOMER baru sekaligus satu rekening SAVINGS.
func (c *authController) Register(ctx context.Context, req RegisterRequest, ip string) (*AuthResponse, error) {
	existing, err := c.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if existing != nil {
		return nil, apperr.ErrDuplicateEmail
	}

	hashed, err := password.Hash(req.Password)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}

	user := &models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashed,
		Role:     models.RoleCustomer,
		Status:   models.UserStatusActive,
	}
	if err := c.userRepo.Create(ctx, user); err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}

	number, err := c.generateAccountNumber(ctx)
	if err != nil {
		return nil, err
	}
	account := &models.Account{
		UserID:        user.ID,
		AccountNumber: number,
		AccountType:   models.AccountTypeSavings,
		Balance:       0,
		Currency:      "IDR",
		Status:        models.AccountStatusActive,
	}
	if err := c.accountRepo.Create(ctx, account); err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}

	recordAudit(ctx, c.auditRepo, uintPtr(user.ID), models.ActionRegister, "Registrasi user baru", ip)
	return c.buildAuthResponse(user)
}

// Login memverifikasi kredensial user dan menghasilkan token akses dan refresh.
func (c *authController) Login(ctx context.Context, req LoginRequest, ip string) (*AuthResponse, error) {
	user, err := c.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if user == nil || !password.Compare(user.Password, req.Password) {
		return nil, apperr.ErrInvalidCredentials
	}
	if !user.IsActive() {
		return nil, apperr.ErrUserBlocked
	}

	recordAudit(ctx, c.auditRepo, uintPtr(user.ID), models.ActionLogin, "User login", ip)
	return c.buildAuthResponse(user)
}

// Refresh memverifikasi token refresh dan menghasilkan token akses baru.
func (c *authController) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	claims, err := c.jwtManager.Parse(req.RefreshToken)
	if err != nil || claims.Type != jwtpkg.TokenTypeRefresh {
		return nil, apperr.ErrUnauthorized
	}
	if c.blacklist.IsRevoked(claims.ID) {
		return nil, apperr.ErrUnauthorized
	}

	user, err := c.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}
	if !user.IsActive() {
		return nil, apperr.ErrUserBlocked
	}

	// Refresh token lama dicabut (rotasi token).
	if claims.ExpiresAt != nil {
		c.blacklist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}
	return c.buildAuthResponse(user)
}

// Logout mencabut token akses dan refresh yang diberikan.
func (c *authController) Logout(ctx context.Context, claims *jwtpkg.Claims, ip string) error {
	if claims == nil {
		return apperr.ErrUnauthorized
	}
	if claims.ExpiresAt != nil {
		c.blacklist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}
	recordAudit(ctx, c.auditRepo, uintPtr(claims.UserID), models.ActionLogout, "User logout", ip)
	return nil
}

// buildAuthResponse membangun respons autentikasi dengan token akses, token refresh, dan informasi user.
func (c *authController) buildAuthResponse(user *models.User) (*AuthResponse, error) {
	accessToken, expiresAt, err := c.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	refreshToken, _, err := c.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    expiresAt,
		User:         NewUserResponse(user),
	}, nil
}

// generateAccountNumber mencoba beberapa kali sampai mendapat nomor unik.
func (c *authController) generateAccountNumber(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		number := utils.GenerateAccountNumber()
		exists, err := c.accountRepo.ExistsByNumber(ctx, number)
		if err != nil {
			return "", apperr.ErrInternal.Wrap(err)
		}
		if !exists {
			return number, nil
		}
	}
	return "", apperr.ErrDuplicateAccount
}
