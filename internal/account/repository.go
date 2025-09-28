package account

import (
	"context"
	"spsyncpro_api/pkg/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gorm.io/gorm"
)

type AccountRepo struct {
	db    *gorm.DB
	trace trace.Tracer
}

func NewAccountRepository(db *gorm.DB) domain.AccountRepository {
	trace := otel.Tracer("accountRepository")
	return &AccountRepo{
		db:    db,
		trace: trace,
	}
}

func (r *AccountRepo) CreateAccount(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	_, span := r.trace.Start(ctx, "CreateAccount")
	defer span.End()
	err := r.db.Create(account).Error
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepo) GetAccountByEmail(ctx context.Context, email string) (*domain.Account, error) {
	_, span := r.trace.Start(ctx, "GetAccountByEmail")
	defer span.End()
	var account domain.Account
	err := r.db.Where("email = ?", email).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepo) GetAccountByID(ctx context.Context, id uint) (*domain.Account, error) {
	_, span := r.trace.Start(ctx, "GetAccountByID")
	defer span.End()
	var account domain.Account
	err := r.db.Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepo) UpdateAccount(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	_, span := r.trace.Start(ctx, "UpdateAccount")
	defer span.End()
	err := r.db.Save(account).Error
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepo) DeleteAccount(ctx context.Context, id uint) error {
	_, span := r.trace.Start(ctx, "DeleteAccount")
	defer span.End()
	return r.db.Delete(&domain.Account{}, id).Error
}

func (r *AccountRepo) LogAccountActivity(ctx context.Context, accountID uint, activity string) error {
	_, span := r.trace.Start(ctx, "LogAccountActivity")
	defer span.End()
	return r.db.Create(&domain.AccountActivity{AccountID: accountID, Activity: activity}).Error
}
