package account_test

import (
	"context"
	"go_starter_api/internal/account"
	"go_starter_api/pkg/domain"
	"go_starter_api/pkg/mailer"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestAccountService_HashPassword(t *testing.T) {

	otel.SetTracerProvider(noop.NewTracerProvider())

	emailService := mailer.NewMockEmailService(t)
	t.Run("should hash and compare password correctly", func(t *testing.T) {
		service := account.NewAccountService(emailService)

		password := "password"
		hash, err := service.HashPassword(context.Background(), password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)

		ok, err := service.ComparePassword(context.Background(), password, hash)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("should return error if password is empty", func(t *testing.T) {
		service := account.NewAccountService(nil)

		password := ""
		hash, err := service.HashPassword(context.Background(), password)
		assert.ErrorIs(t, err, domain.ErrPasswordEmpty)
		assert.Empty(t, hash)
	})
}

func TestAccountService_GenerateAndValidateToken(t *testing.T) {
	// Set up test environment
	viper.Set("JWT_SECRET", "test_secret_key_for_jwt_validation")
	defer viper.Reset()

	emailService := mailer.NewMockEmailService(t)
	service := account.NewAccountService(emailService)

	t.Run("should generate and validate token correctly", func(t *testing.T) {
		account := &domain.Account{ID: 123, Email: "test@example.com"}

		// Generate token
		token, err := service.GenerateAuthToken(context.Background(), account)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate token
		accountID, err := service.ValidateAuthToken(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, uint(123), accountID)
	})

	t.Run("should return error if JWT secret is not set", func(t *testing.T) {
		// Temporarily unset JWT secret
		viper.Set("JWT_SECRET", "")

		account := &domain.Account{ID: 1, Email: "test@test.com"}
		token, err := service.GenerateAuthToken(context.Background(), account)
		assert.Error(t, err)
		assert.Empty(t, token)
	})

	t.Run("should return error if token is invalid", func(t *testing.T) {
		invalidToken := "invalid_token"
		accountID, err := service.ValidateAuthToken(context.Background(), invalidToken)
		assert.Error(t, err)
		assert.Equal(t, uint(0), accountID)
	})

	t.Run("should return error if token is malformed", func(t *testing.T) {
		malformedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid"
		accountID, err := service.ValidateAuthToken(context.Background(), malformedToken)
		assert.Error(t, err)
		assert.Equal(t, uint(0), accountID)
	})
}

func TestAccountService_GenerateAndValidatePasswordResetToken(t *testing.T) {
	viper.Set("JWT_SECRET", "test_secret_key_for_jwt_validation")
	defer viper.Reset()

	emailService := mailer.NewMockEmailService(t)
	service := account.NewAccountService(emailService)

	t.Run("should generate and validate password reset token correctly", func(t *testing.T) {
		account := &domain.Account{ID: 123, Email: "test@example.com"}

		// Generate token
		token, err := service.GeneratePasswordResetToken(context.Background(), account)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate token
		accountID, err := service.ValidatePasswordResetToken(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, uint(123), accountID)
	})

	t.Run("should return error if JWT secret is not set", func(t *testing.T) {
		viper.Set("JWT_SECRET", "")
		defer viper.Reset()

		account := &domain.Account{ID: 1, Email: "test@test.com"}
		token, err := service.GeneratePasswordResetToken(context.Background(), account)
		assert.Error(t, err)
		assert.Empty(t, token)
	})

	t.Run("should return error if token is invalid", func(t *testing.T) {
		invalidToken := "invalid_token"
		accountID, err := service.ValidatePasswordResetToken(context.Background(), invalidToken)
		assert.Error(t, err)
		assert.Equal(t, uint(0), accountID)
	})

	t.Run("should return error if token is malformed", func(t *testing.T) {
		malformedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid"
		accountID, err := service.ValidatePasswordResetToken(context.Background(), malformedToken)
		assert.Error(t, err)
		assert.Equal(t, uint(0), accountID)
	})
}

func TestAccountService_SendPasswordResetEmail(t *testing.T) {

	t.Run("should send password reset email correctly", func(t *testing.T) {
		viper.Set("SERVER_URL", "http://localhost:8080")
		defer viper.Reset()

		emailService := mailer.NewMockEmailService(t)
		// Set up the mock to expect SendEmail to be called with the correct arguments
		emailService.
			On(
				"SendEmail",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
			).
			Return(nil).
			Once()

		service := account.NewAccountService(emailService)

		email := "test@example.com"
		token := "test_token"
		err := service.SendPasswordResetEmail(context.Background(), email, token)
		assert.NoError(t, err)
	})

	t.Run("should return error if server url is not set", func(t *testing.T) {
		viper.Set("SERVER_URL", "")
		defer viper.Reset()

		emailService := mailer.NewMockEmailService(t)
		service := account.NewAccountService(emailService)

		email := "test@example.com"
		token := "test_token"
		err := service.SendPasswordResetEmail(context.Background(), email, token)
		assert.ErrorIs(t, err, domain.ErrServerURLNotSet)
	})

}
