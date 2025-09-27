package mailer

import (
	"net/smtp"

	"github.com/spf13/viper"
)

type EmailService interface {
	SendEmail(email string, subject string, body string) error
}

type EmailServiceImpl struct {
	user     string
	password string
	smtpHost string
	smtpPort string
	smtpFrom string
}

func NewEmailService() EmailService {
	return &EmailServiceImpl{
		user:     viper.GetString("SMTP_USER"),
		password: viper.GetString("SMTP_PASSWORD"),
		smtpHost: viper.GetString("SMTP_HOST"),
		smtpPort: viper.GetString("SMTP_PORT"),
		smtpFrom: viper.GetString("SMTP_FROM"),
	}
}

func (e *EmailServiceImpl) SendEmail(email string, subject string, body string) error {
	// use nil auth if user and password are not set
	var auth smtp.Auth

	if viper.GetString("GIN_MODE") != "release" {
		auth = nil
	} else {
		auth = smtp.PlainAuth("", e.user, e.password, e.smtpHost)
	}

	msg := []byte("To: " + email + "\r\n" + "Subject: " + subject + "\r\n" + "\r\n" + body)

	err := smtp.SendMail(e.smtpHost+":"+e.smtpPort, auth, e.smtpFrom, []string{email}, msg)
	if err != nil {
		return err
	}

	return nil
}
