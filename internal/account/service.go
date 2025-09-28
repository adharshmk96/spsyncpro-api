package account

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"spsyncpro_api/pkg/domain"
	"spsyncpro_api/pkg/mailer"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/argon2"
)

var (
	ErrFailedToGenerateSalt = errors.New("failed to generate salt")
	ErrJWTSecretNotSet      = errors.New("jwt secret is not set")
	ErrSubjectClaimNotFound = errors.New("subject claim not found in token")
	ErrInvalidSubjectClaim  = errors.New("invalid subject claim type")
)

type AccountService struct {
	tracer       trace.Tracer
	emailService mailer.EmailService
}

func NewAccountService(emailService mailer.EmailService) domain.AccountService {
	tracer := otel.Tracer("accountService")
	return &AccountService{
		tracer:       tracer,
		emailService: emailService,
	}
}

// hashes password into following format:
// $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
func (s *AccountService) HashPassword(ctx context.Context, password string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "HashPassword")
	defer span.End()

	// Validate input type and length
	if len(password) == 0 {
		return "", domain.ErrPasswordEmpty
	}

	// Argon2id parameters
	var (
		memory  uint32 = 64 * 1024 // 64 MB
		time    uint32 = 1         // 1 iteration
		threads uint8  = 4         // 4 threads
		keyLen  uint32 = 32        // 32 bytes
		saltLen int    = 16        // 16 bytes
	)

	// Generate a random salt
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToGenerateSalt, err)
	}

	// Hash the password using Argon2id
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Encode salt and hash to base64 for storage
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, time, threads, b64Salt, b64Hash)

	return encoded, nil
}

func (s *AccountService) ComparePassword(ctx context.Context, password, hash string) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "ComparePassword")
	defer span.End()

	// Split the hash into its components
	// Expected format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, domain.ErrInvalidHashFormat
	}

	// Validate the algorithm and version
	if parts[1] != "argon2id" {
		return false, domain.ErrInvalidHashFormat
	}

	// Parse the parameters from the third part: m=65536,t=1,p=4
	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, domain.ErrInvalidHashFormat
	}

	// Extract memory parameter
	memoryStr := strings.TrimPrefix(params[0], "m=")
	memory, err := strconv.ParseUint(memoryStr, 10, 32)
	if err != nil {
		return false, domain.ErrInvalidHashFormat
	}

	// Extract time parameter
	timeStr := strings.TrimPrefix(params[1], "t=")
	time, err := strconv.ParseUint(timeStr, 10, 32)
	if err != nil {
		return false, domain.ErrInvalidHashFormat
	}

	// Extract threads parameter
	threadsStr := strings.TrimPrefix(params[2], "p=")
	threads, err := strconv.ParseUint(threadsStr, 10, 32)
	if err != nil {
		return false, domain.ErrInvalidHashFormat
	}

	// Extract the salt and hash (parts[4] and parts[5])
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, domain.ErrInvalidHashFormat
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, domain.ErrInvalidHashFormat
	}

	// Use the same keyLen as in HashPassword (32 bytes)
	keyLen := uint32(32)

	// Verify the password
	computedHash := argon2.IDKey([]byte(password), salt, uint32(time), uint32(memory), uint8(threads), keyLen)

	// Compare the computed hash with the stored hash
	return hmac.Equal(hashBytes, computedHash), nil
}

func (s *AccountService) GenerateAuthToken(ctx context.Context, account *domain.Account) (string, error) {
	ctx, span := s.tracer.Start(ctx, "GenerateAuthToken")
	defer span.End()

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		return "", ErrJWTSecretNotSet
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": account.ID,
		"iss": "spsyncpro_api",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(jwtSecret))
}

func (s *AccountService) ValidateAuthToken(ctx context.Context, token string) (uint, error) {
	ctx, span := s.tracer.Start(ctx, "ValidateAuthToken")
	defer span.End()

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		return 0, ErrJWTSecretNotSet
	}

	claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	// Extract the subject claim and convert from float64 (JSON number) to uint
	subClaim, ok := claims.Claims.(jwt.MapClaims)["sub"]
	if !ok {
		return 0, ErrSubjectClaimNotFound
	}

	// Convert float64 to uint (JWT library returns JSON numbers as float64)
	accountIDFloat, ok := subClaim.(float64)
	if !ok {
		return 0, ErrInvalidSubjectClaim
	}

	return uint(accountIDFloat), nil
}

func (s *AccountService) GeneratePasswordResetToken(ctx context.Context, account *domain.Account) (string, error) {
	ctx, span := s.tracer.Start(ctx, "GeneratePasswordResetToken")
	defer span.End()

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		return "", ErrJWTSecretNotSet
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(account.ID), 10) + ":password-reset",
		"iss": "spsyncpro_api",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(jwtSecret))
}

func (s *AccountService) ValidatePasswordResetToken(ctx context.Context, token string) (uint, error) {
	ctx, span := s.tracer.Start(ctx, "ValidatePasswordResetToken")
	defer span.End()

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		return 0, ErrJWTSecretNotSet
	}

	claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	subClaim, ok := claims.Claims.(jwt.MapClaims)["sub"]
	if !ok {
		return 0, ErrSubjectClaimNotFound
	}

	parts := strings.Split(subClaim.(string), ":")
	if len(parts) != 2 {
		return 0, ErrInvalidSubjectClaim
	}

	accountID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	if parts[1] != "password-reset" {
		return 0, ErrInvalidSubjectClaim
	}

	return uint(accountID), nil
}

func (s *AccountService) SendPasswordResetEmail(ctx context.Context, email string, token string) error {
	ctx, span := s.tracer.Start(ctx, "SendPasswordResetEmail")
	defer span.End()

	serverUrl := viper.GetString("SERVER_URL")
	if serverUrl == "" {
		return domain.ErrServerURLNotSet
	}
	link := serverUrl + "/api/v1/account/reset-password?token=" + token

	resetPasswordTemplate := `
		<html>
		<body>
			<h1>Password Reset Request</h1>
			<p><a href="` + link + `">Click here to reset your password</a></p>
			<p>If you did not request a password reset, please ignore this email.</p>
			<p>Thank you for using our service.</p>
		</body>
		</html>
	`

	return s.emailService.SendEmail(email, "Password Reset", resetPasswordTemplate)
}
