package domain

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Email     string         `json:"email" gorm:"unique"`
	Password  string         `json:"password"`
}

var (
	ActivityLogin          = "login"
	ActivityLogout         = "logout"
	ActivityRegister       = "register"
	ActivityUpdate         = "update"
	ActivityDelete         = "delete"
	ActivityResetPassword  = "reset_password"
	ActivityForgotPassword = "forgot_password"
	ActivityChangePassword = "change_password"
)

type AccountActivity struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	AccountID uint   `json:"account_id"`
	Activity  string `json:"activity"`
}

type AccountService interface {
	GenerateAuthToken(ctx context.Context, account *Account) (string, error)
	ValidateAuthToken(ctx context.Context, token string) (uint, error)
	HashPassword(ctx context.Context, password string) (string, error)
	ComparePassword(ctx context.Context, password, hash string) (bool, error)

	GeneratePasswordResetToken(ctx context.Context, account *Account) (string, error)
	ValidatePasswordResetToken(ctx context.Context, token string) (uint, error)
	SendPasswordResetEmail(ctx context.Context, email string, token string) error
}

var (
	ErrPasswordEmpty     = errors.New("password cannot be empty")
	ErrInvalidHashFormat = errors.New("invalid hash format")
	ErrServerURLNotSet   = errors.New("server url is not set")
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account *Account) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id uint) (*Account, error)
	UpdateAccount(ctx context.Context, account *Account) (*Account, error)
	DeleteAccount(ctx context.Context, id uint) error

	LogAccountActivity(ctx context.Context, accountID uint, activity string) error
}
