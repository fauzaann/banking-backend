package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

// AppError adalah error domain yang membawa HTTP status dan error code.
type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"error"`
	Message string `json:"message"`
	err     error
}

func (e *AppError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.err }

// Wrap membungkus error teknis di balik AppError tanpa mengubah pesan publik.
func (e *AppError) Wrap(err error) *AppError {
	clone := *e
	clone.err = err
	return &clone
}

func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// As mengembalikan *AppError dari error apa pun; fallback ke internal error.
func As(err error) *AppError {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr
	}
	return ErrInternal.Wrap(err)
}

var (
	ErrUserNotFound        = New(http.StatusNotFound, "USER_NOT_FOUND", "User tidak ditemukan")
	ErrAccountNotFound     = New(http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Rekening tidak ditemukan")
	ErrTransactionNotFound = New(http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaksi tidak ditemukan")
	ErrTransferNotFound    = New(http.StatusNotFound, "TRANSFER_NOT_FOUND", "Transfer tidak ditemukan")
	ErrBeneficiaryNotFound = New(http.StatusNotFound, "BENEFICIARY_NOT_FOUND", "Beneficiary tidak ditemukan")

	ErrInvalidCredentials = New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email atau password salah")
	ErrUnauthorized       = New(http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau sudah kedaluwarsa")
	ErrForbidden          = New(http.StatusForbidden, "FORBIDDEN", "Anda tidak memiliki akses ke resource ini")
	ErrUserBlocked        = New(http.StatusForbidden, "USER_BLOCKED", "Akun Anda diblokir, hubungi administrator")

	ErrInsufficientBalance = New(http.StatusUnprocessableEntity, "INSUFFICIENT_BALANCE", "Saldo tidak mencukupi")
	ErrAccountFrozen       = New(http.StatusUnprocessableEntity, "ACCOUNT_FROZEN", "Rekening sedang dibekukan")
	ErrAccountClosed       = New(http.StatusUnprocessableEntity, "ACCOUNT_CLOSED", "Rekening sudah ditutup")
	ErrSelfTransfer        = New(http.StatusUnprocessableEntity, "SELF_TRANSFER", "Tidak dapat transfer ke rekening sendiri")
	ErrInvalidAmount       = New(http.StatusUnprocessableEntity, "INVALID_AMOUNT", "Nominal harus lebih besar dari 0")
	ErrTransferFailed      = New(http.StatusUnprocessableEntity, "TRANSFER_FAILED", "Transfer gagal diproses")
	ErrWrongPassword       = New(http.StatusUnprocessableEntity, "WRONG_PASSWORD", "Password lama tidak sesuai")

	ErrDuplicateEmail   = New(http.StatusConflict, "DUPLICATE_EMAIL", "Email sudah terdaftar")
	ErrDuplicateAccount = New(http.StatusConflict, "DUPLICATE_ACCOUNT", "Nomor rekening sudah terdaftar")
	ErrDuplicateBenef   = New(http.StatusConflict, "DUPLICATE_BENEFICIARY", "Rekening sudah ada di daftar beneficiary")

	ErrBadRequest   = New(http.StatusBadRequest, "BAD_REQUEST", "Format request tidak valid")
	ErrTooManyReq   = New(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Terlalu banyak permintaan, coba lagi nanti")
	ErrInternal     = New(http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server")
	ErrNotImplement = New(http.StatusNotImplemented, "NOT_IMPLEMENTED", "Fitur belum tersedia")
)
