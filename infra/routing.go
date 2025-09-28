package infra

import (
	"spsyncpro_api/internal/account"
	"spsyncpro_api/internal/organization"
	"spsyncpro_api/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SetupRoutes(
	rg *gin.RouterGroup,
	db *gorm.DB,
	logger *logrus.Logger,
) {
	emailService := mailer.NewEmailService()

	accountRepository := account.NewAccountRepository(db)
	accountService := account.NewAccountService(emailService)
	accountHandler := account.NewAccountHandler(logger, accountService, accountRepository)

	rg.POST("/account/register", accountHandler.RegisterAccount)
	rg.POST("/account/login", accountHandler.LoginAccount)
	rg.POST("/account/forgot-password", accountHandler.ForgotPassword)
	rg.POST("/account/reset-password", accountHandler.ResetPassword)

	rg.Use(account.AuthMiddleware(accountService))

	rg.GET("/account/profile", accountHandler.GetProfile)
	rg.POST("/account/logout", accountHandler.LogoutAccount)
	rg.POST("/account/change-password", accountHandler.ChangePassword)

	organizationRepository := organization.NewOrganizationRepository(db)
	organizationService := organization.NewOrganizationService()
	organizationHandler := organization.NewOrganizationHandler(organizationService, organizationRepository)

	rg.POST("/organization/upsert", organizationHandler.UpsertOrganization)
	rg.GET("/organization/get", organizationHandler.GetOrganization)
	rg.DELETE("/organization/delete", organizationHandler.DeleteOrganization)
	rg.GET("/organization/check-authorization", organizationHandler.CheckAuthorization)
}
