package account

import (
	"errors"
	"net/http"
	"spsyncpro_api/pkg/domain"
	"spsyncpro_api/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type AccountHandler struct {
	logger *logrus.Logger
	tracer trace.Tracer
	meter  metric.Meter

	accountService    domain.AccountService
	accountRepository domain.AccountRepository
}

const (
	name = "accountHandler"
)

func NewAccountHandler(
	logger *logrus.Logger,
	accountService domain.AccountService,
	accountRepository domain.AccountRepository,
) *AccountHandler {
	tracer := otel.Tracer(name)
	meter := otel.Meter(name)
	return &AccountHandler{
		logger:            logger,
		tracer:            tracer,
		meter:             meter,
		accountService:    accountService,
		accountRepository: accountRepository,
	}
}

type RegisterAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterAccountResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

// @Summary		Register a new account
// @Description	Register a new account
// @Tags			account
// @Accept			json
// @Produce		json
// @Param			account	body		RegisterAccountRequest	true	"Account"
// @Success		200		{object}	RegisterAccountResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/register [post]
func (h *AccountHandler) RegisterAccount(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "RegisterAccount")
	defer span.End()

	var req RegisterAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if account already exists
	existingAcc, err := h.accountRepository.GetAccountByEmail(ctx, req.Email)
	if err == nil && existingAcc != nil {
		h.logger.WithField("userId", existingAcc.ID).Errorf("account already exists")
		c.JSON(http.StatusBadRequest, gin.H{"error": "account already exists"})
		return
	}
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Errorf("failed to get account by email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	// Hash the password before storing
	hashedPassword, err := h.accountService.HashPassword(ctx, req.Password)
	if err != nil {
		h.logger.Errorf("failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	acc := &domain.Account{
		Email:    req.Email,
		Password: hashedPassword,
	}

	acc, err = h.accountRepository.CreateAccount(ctx, acc)
	if err != nil {
		h.logger.Errorf("failed to create account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := h.accountService.GenerateAuthToken(ctx, acc)
	if err != nil {
		h.logger.WithField("userId", acc.ID).Errorf("failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.accountRepository.LogAccountActivity(ctx, acc.ID, domain.ActivityRegister)
	if err != nil {
		h.logger.WithField("userId", acc.ID).Errorf("failed to log activity: %v", err)
	}

	c.JSON(http.StatusOK, RegisterAccountResponse{
		ID:    acc.ID,
		Email: acc.Email,
		Token: token,
	})
}

type LoginAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginAccountResponse struct {
	Token string `json:"token"`
}

// @Summary		Login a user
// @Description	Login a user
// @Tags			account
// @Accept			json
// @Produce		json
// @Param			account	body		LoginAccountRequest	true	"Account"
// @Success		200		{object}	LoginAccountResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/login [post]
func (h *AccountHandler) LoginAccount(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "LoginAccount")
	defer span.End()

	var req LoginAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.accountRepository.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.WithField("email", req.Email).Errorf("account not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		}
		h.logger.Errorf("failed to get account by email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ok, err := h.accountService.ComparePassword(ctx, req.Password, acc.Password)
	if err != nil {
		h.logger.WithField("userId", acc.ID).Errorf("failed to compare password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !ok {
		h.logger.WithField("userId", acc.ID).Errorf("invalid password")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.accountService.GenerateAuthToken(ctx, acc)
	if err != nil {
		h.logger.WithField("userId", acc.ID).Errorf("failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	err = h.accountRepository.LogAccountActivity(ctx, acc.ID, domain.ActivityLogin)
	if err != nil {
		h.logger.WithField("userId", acc.ID).Errorf("failed to log activity: %v", err)
	}

	c.JSON(
		http.StatusOK,
		LoginAccountResponse{
			Token: token,
		},
	)
}

// @Summary		Logout a user
// @Description	Logout a user
// @Tags			account
// @Accept			json
// @Produce		json
// @Success		200		{object}	map[string]string
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/logout [post]
func (h *AccountHandler) LogoutAccount(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "LogoutAccount")
	defer span.End()

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		h.logger.Errorf("accountID not found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err := h.accountRepository.LogAccountActivity(ctx, accountID, domain.ActivityLogout)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to log activity: %v", err)
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "logout successful",
		},
	)
}

type GetProfileResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// @Summary		Get Profile
// @Description	Get Profile of the authenticated user
// @Tags			account
// @Accept			json
// @Produce		json
// @Success		200		{object}	GetProfileResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/profile [get]
func (h *AccountHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "GetProfile")
	defer span.End()

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		h.logger.Errorf("accountID not found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	acc, err := h.accountRepository.GetAccountByID(ctx, accountID)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to get account by id: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, GetProfileResponse{
		ID:        acc.ID,
		Email:     acc.Email,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
	})
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// @Summary		Forgot Password
// @Description	Forgot Password
// @Tags			account
// @Accept			json
// @Produce		json
// @Param			account	body		ForgotPasswordRequest	true	"Account"
// @Success		200		{object}	ForgotPasswordResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/forgot-password [post]
func (h *AccountHandler) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "ForgotPassword")
	defer span.End()

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.accountRepository.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Errorf("failed to get account by email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if acc == nil {
		h.logger.Errorf("account not found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "account not found"})
		return
	}

	token, err := h.accountService.GeneratePasswordResetToken(ctx, acc)
	if err != nil {
		h.logger.Errorf("failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	err = h.accountService.SendPasswordResetEmail(ctx, acc.Email, token)
	if err != nil {
		h.logger.Errorf("failed to send password reset email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send password reset email"})
		return
	}

	err = h.accountRepository.LogAccountActivity(ctx, acc.ID, domain.ActivityForgotPassword)
	if err != nil {
		h.logger.Errorf("failed to log activity: %v", err)
	}

	c.JSON(
		http.StatusOK,
		ForgotPasswordResponse{
			Message: "password reset email sent",
		},
	)
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// @Summary		Reset Password
// @Description	Reset Password
// @Tags			account
// @Accept			json
// @Produce		json
// @Param			account	body		ResetPasswordRequest	true	"Account"
// @Success		200		{object}	ResetPasswordResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/reset-password [post]
func (h *AccountHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "ResetPassword")
	defer span.End()

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := req.Token
	password := req.Password

	accountID, err := h.accountService.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		h.logger.Errorf("failed to validate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	acc, err := h.accountRepository.GetAccountByID(ctx, accountID)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to get account by id: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	hashedPassword, err := h.accountService.HashPassword(ctx, password)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	acc.Password = hashedPassword

	acc, err = h.accountRepository.UpdateAccount(ctx, acc)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to update account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = h.accountRepository.LogAccountActivity(ctx, acc.ID, domain.ActivityResetPassword)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to log activity: %v", err)
	}

	c.JSON(
		http.StatusOK,
		ResetPasswordResponse{
			Message: "password reset successful",
		},
	)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// @Summary		Change Password
// @Description	Change Password
// @Tags			account
// @Accept			json
// @Produce		json
// @Param			account	body		ChangePasswordRequest	true	"Account"
// @Success		200		{object}	ChangePasswordResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/account/change-password [post]
func (h *AccountHandler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "ChangePassword")
	defer span.End()

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		h.logger.Errorf("accountID not found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	acc, err := h.accountRepository.GetAccountByID(ctx, accountID)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to get account by id: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ok, err := h.accountService.ComparePassword(ctx, req.OldPassword, acc.Password)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to compare password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !ok {
		h.logger.WithField("userId", accountID).Errorf("invalid old password")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid old password"})
		return
	}

	hashedPassword, err := h.accountService.HashPassword(ctx, req.NewPassword)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	acc.Password = hashedPassword

	acc, err = h.accountRepository.UpdateAccount(ctx, acc)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to update account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = h.accountRepository.LogAccountActivity(ctx, acc.ID, domain.ActivityChangePassword)
	if err != nil {
		h.logger.WithField("userId", accountID).Errorf("failed to log activity: %v", err)
	}

	c.JSON(
		http.StatusOK,
		ChangePasswordResponse{
			Message: "password changed successfully",
		},
	)
}
