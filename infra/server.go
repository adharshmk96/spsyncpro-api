package infra

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Config struct {
	Port int
}

func ginServerMode() string {
	if viper.GetString("SERVER_MODE") != "production" {
		return gin.DebugMode
	}
	return gin.ReleaseMode
}

func NewServer(
	db *gorm.DB,
	logger *logrus.Logger,
	config Config,
) *http.Server {
	gin.SetMode(ginServerMode())

	router := gin.Default()
	router.Use(otelgin.Middleware("go_starter-api"))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	rg := router.Group("/api/v1")

	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	SetupRoutes(rg, db, logger)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	return srv
}
