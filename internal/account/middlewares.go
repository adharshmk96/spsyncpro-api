package account

import (
	"net/http"
	"spsyncpro_api/pkg/domain"
	"spsyncpro_api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const AuthHeaderKey = "Authorization"

func AuthMiddleware(accountService domain.AccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(AuthHeaderKey)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		accountID, err := accountService.ValidateAuthToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set(utils.AccountIdContextKey, accountID)

		c.Next()
	}
}
