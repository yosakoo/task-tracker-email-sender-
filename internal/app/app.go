package app

import (
	"flag"

	"github.com/yosakoo/task-tracker-email-sender-/internal/config"
	"github.com/yosakoo/task-tracker-email-sender-/internal/services"
	"github.com/yosakoo/task-tracker-email-sender-/pkg/logger"
	"github.com/yosakoo/task-tracker-email-sender-/internal/delivery/rabbitmq"
)

func Run(cfg *config.Config) {
	lifetime := flag.Duration("lifetime", 0, "Lifetime for the consumer (0 to run until manually stopped)")
	flag.Parse()	
	l := logger.New(cfg.Log.Level)
    l.Info("start server")

	emailService := services.NewEmailService(l,services.Config{
		From: cfg.SmtpServer.From,
		Password: cfg.SmtpServer.Password,
		SmtpServer: cfg.SmtpServer.Server,
		Port: cfg.SmtpServer.Port,
	})

	rmqCfg := rabbitmq.Config{
		URL:          cfg.RabbitMQ.URL,
		Exchange:     cfg.RabbitMQ.Exchange,
		ExchangeType: cfg.RabbitMQ.ExchangeType,
		Queue:        cfg.RabbitMQ.Queue,
		BindingKey:   "",
		ConsumerTag:  cfg.RabbitMQ.ConsumerTag,
	}
	rabbitmq.New(rmqCfg, lifetime, emailService)
}
