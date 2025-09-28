package infra

import (
	"fmt"
	"spsyncpro_api/pkg/domain"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitGormDB() *gorm.DB {
	var db *gorm.DB
	var err error

	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")
	dbname := viper.GetString("DB_NAME")
	sslmode := viper.GetString("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	timezone := viper.GetString("DB_TIMEZONE")
	if timezone == "" {
		timezone = "UTC"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=%s", host, port, user, password, dbname, sslmode, timezone)

	db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&domain.Account{})
	db.AutoMigrate(&domain.AccountActivity{})

	return db
}
