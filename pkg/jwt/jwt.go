package jwt

import (
	"errors"
	"sync"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"banking/pkg/utils"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var ErrInvalidToken = errors.New("token tidak valid")

// Claims menyimpan informasi yang akan disematkan dalam token JWT.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	gojwt.RegisteredClaims
}

// Manager bertanggung jawab untuk membuat dan memvalidasi token JWT.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager membuat instance baru dari Manager dengan secret dan durasi token yang diberikan.
func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// generate membuat token JWT dengan claims yang diberikan dan mengembalikan token string beserta waktu kedaluwarsanya.
func (m *Manager) generate(userID uint, email, role, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        utils.GenerateReferenceNumber("JTI"),
			Subject:   email,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
		},
	}
	token, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// GenerateAccessToken membuat token akses JWT untuk user dengan ID, email, dan role yang diberikan.
func (m *Manager) GenerateAccessToken(userID uint, email, role string) (string, time.Time, error) {
	return m.generate(userID, email, role, TokenTypeAccess, m.accessTTL)
}

// GenerateRefreshToken membuat token refresh JWT untuk user dengan ID, email, dan role yang diberikan.
func (m *Manager) GenerateRefreshToken(userID uint, email, role string) (string, time.Time, error) {
	return m.generate(userID, email, role, TokenTypeRefresh, m.refreshTTL)
}

// Parse memvalidasi signature, algoritma, dan masa berlaku token.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := gojwt.ParseWithClaims(tokenString, claims, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// Blacklist menyimpan JTI token yang sudah di-logout (in-memory).
// Untuk multi-instance production, ganti implementasi ini dengan Redis.
type Blacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

// NewBlacklist membuat instance baru dari Blacklist.
func NewBlacklist() *Blacklist {
	b := &Blacklist{tokens: make(map[string]time.Time)}
	go b.cleanupLoop()
	return b
}

// Revoke menambahkan JTI token ke blacklist dengan waktu kedaluwarsa yang diberikan.
func (b *Blacklist) Revoke(jti string, expiresAt time.Time) {
	if jti == "" {
		return
	}
	b.mu.Lock()
	b.tokens[jti] = expiresAt
	b.mu.Unlock()
}

// IsRevoked memeriksa apakah JTI token ada di blacklist dan belum kedaluwarsa.
func (b *Blacklist) IsRevoked(jti string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	exp, ok := b.tokens[jti]
	return ok && time.Now().Before(exp)
}

// cleanupLoop secara periodik membersihkan JTI token yang sudah kedaluwarsa dari blacklist.
func (b *Blacklist) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		b.mu.Lock()
		for jti, exp := range b.tokens {
			if now.After(exp) {
				delete(b.tokens, jti)
			}
		}
		b.mu.Unlock()
	}
}


