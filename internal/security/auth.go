package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/notemg/notemg/internal/config"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked, try again later")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type Claims struct {
	jwt.RegisteredClaims
	Type string `json:"type"`
}

type Auth struct {
	cfg        *config.Config
	mu         sync.Mutex
	attempts   map[string]*loginAttempt
}

type loginAttempt struct {
	count    int
	lockedAt time.Time
}

func NewAuth(cfg *config.Config) *Auth {
	return &Auth{
		cfg:      cfg,
		attempts: make(map[string]*loginAttempt),
	}
}

func (a *Auth) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), a.cfg.Auth.BcryptCost)
	return string(bytes), err
}

func (a *Auth) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a *Auth) GenerateToken(userID string) (string, string, error) {
	now := time.Now()

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(a.cfg.Auth.TokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "notemg",
		},
		Type: "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(a.cfg.Auth.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(a.cfg.Auth.RefreshExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "notemg",
		},
		Type: "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(a.cfg.Auth.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return accessStr, refreshStr, nil
}

func (a *Auth) ValidateToken(tokenStr string, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(a.cfg.Auth.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Type != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (a *Auth) CheckLoginAttempts(ip string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	attempt, exists := a.attempts[ip]
	if !exists {
		return nil
	}

	if attempt.count >= a.cfg.Auth.MaxLoginAttempts {
		if time.Since(attempt.lockedAt) < a.cfg.Auth.LockDuration {
			return ErrAccountLocked
		}
		delete(a.attempts, ip)
	}

	return nil
}

func (a *Auth) RecordFailedAttempt(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	attempt, exists := a.attempts[ip]
	if !exists {
		attempt = &loginAttempt{}
		a.attempts[ip] = attempt
	}
	attempt.count++
	if attempt.count >= a.cfg.Auth.MaxLoginAttempts {
		attempt.lockedAt = time.Now()
	}
}

func (a *Auth) ResetLoginAttempts(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.attempts, ip)
}

const passwordFile = ".password_hash"

func (a *Auth) HasPassword(dataDir string) bool {
	path := filepath.Join(dataDir, passwordFile)
	_, err := os.Stat(path)
	return err == nil
}

func (a *Auth) SavePassword(dataDir string, password string) error {
	hash, err := a.HashPassword(password)
	if err != nil {
		return err
	}
	path := filepath.Join(dataDir, passwordFile)
	return os.WriteFile(path, []byte(hash), 0600)
}

func (a *Auth) VerifyPassword(dataDir string, password string) error {
	path := filepath.Join(dataDir, passwordFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return ErrInvalidCredentials
	}
	if !a.CheckPassword(password, string(data)) {
		return ErrInvalidCredentials
	}
	return nil
}
