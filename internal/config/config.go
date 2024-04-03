package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		RabbitMQ            `yaml:"rabbitmq"`
		Log      	        `yaml:"logger"`
		SmtpServer          `yaml:"smtp_server"`
	}
	RabbitMQ struct {
		URL        string   `yaml:"url"`
		Exchange   string   `yaml:"exchange"`
		ExchangeType string `yaml:"exchange_type"`
		Queue string        `yaml:"queue"`
		ConsumerTag string  `yaml:"consumer_tag"`
	}
	Log struct {
		Level string        `yaml:"log_level"`
	}
	SmtpServer struct{
		From       string   
		Password   string  
		Server 	   string   
		Port       int      `yaml:"port"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{
		RabbitMQ: RabbitMQ{
			URL:os.Getenv("RABBITMQ_URL"),
		},
		SmtpServer: SmtpServer{
			From:       os.Getenv("SMTP_FROM"),
			Password:   os.Getenv("SMTP_PASSWORD"),
			Server:     os.Getenv("SMTP_SERVER"),
		},
	}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	fmt.Println(cfg.SmtpServer)

	return cfg, nil
}
