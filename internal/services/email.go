package services

import (
	"fmt"
	"net/smtp"

	"github.com/yosakoo/task-tracker-email-sender-/pkg/logger"
)

type Config struct {
	From       string
	Password   string
	SmtpServer string
	Port       int
}

type Emails interface {
	SendEmail(to, subject, body string) error
}

type EmailService struct {
	Config

	Log *logger.Logger
}

func NewEmailService(log *logger.Logger,cfg Config) *EmailService{
	return &EmailService{
		Config: cfg,
		Log:log,
	}
}

func(es *EmailService) SendEmail(to, subject, body string) error{
	message := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))
	fmt.Println( es.Config.From, es.Config.Password, es.Config.SmtpServer)
	fmt.Println( es.Config.From, es.Config.Password, es.Config.SmtpServer)
	fmt.Println( es.Config.From, es.Config.Password, es.Config.SmtpServer)
	fmt.Println( es.Config.From, es.Config.Password, es.Config.SmtpServer)
	fmt.Println( es.Config.From, es.Config.Password, es.Config.SmtpServer)
	auth := smtp.PlainAuth("", es.Config.From, es.Config.Password, es.Config.SmtpServer)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", es.Config.SmtpServer, es.Config.Port), auth, es.Config.From, []string{to}, message)
	if err != nil {
		es.Log.Error("Ошибка при отправке письма:", err)
		return err
	}

	return nil
}