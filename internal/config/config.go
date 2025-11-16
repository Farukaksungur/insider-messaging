package config

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	WebhookURL            string
	WebhookAuthKey        string
	APIKey                string
	RedisAddr             string
	RedisPassword         string
	MsgCharLimit          int
	ScheduleSec           int
	MsgPerTick            int
	WebhookTimeoutSeconds int
}

// Load environment variable'ları yükler ve config oluşturur
func Load() (*Config, error) {
	_ = godotenv.Load()

	limit := 160
	if v := os.Getenv("MSG_CHAR_LIMIT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			limit = i
		}
	}
	sched := 120
	if v := os.Getenv("SCHEDULE_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			sched = i
		}
	}
	per := 2
	if v := os.Getenv("MSG_PER_TICK"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			per = i
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	webhookTimeout := 30
	if v := os.Getenv("WEBHOOK_TIMEOUT_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			webhookTimeout = i
		}
	}

	cfg := &Config{
		Port:                  port,
		DBHost:                os.Getenv("DB_HOST"),
		DBPort:                os.Getenv("DB_PORT"),
		DBUser:                os.Getenv("DB_USER"),
		DBPassword:            os.Getenv("DB_PASSWORD"),
		DBName:                os.Getenv("DB_NAME"),
		WebhookURL:            os.Getenv("WEBHOOK_URL"),
		WebhookAuthKey:        os.Getenv("WEBHOOK_AUTH_KEY"),
		APIKey:                os.Getenv("API_KEY"),
		RedisAddr:             os.Getenv("REDIS_ADDR"),
		RedisPassword:         os.Getenv("REDIS_PASSWORD"),
		MsgCharLimit:          limit,
		ScheduleSec:           sched,
		MsgPerTick:            per,
		WebhookTimeoutSeconds: webhookTimeout,
	}

	if cfg.DBHost == "" {
		return nil, errors.New("DB_HOST is required")
	}
	if cfg.DBPort == "" {
		return nil, errors.New("DB_PORT is required")
	}
	if cfg.DBUser == "" {
		return nil, errors.New("DB_USER is required")
	}
	if cfg.DBPassword == "" {
		return nil, errors.New("DB_PASSWORD is required")
	}
	if cfg.DBName == "" {
		return nil, errors.New("DB_NAME is required")
	}
	if cfg.WebhookURL == "" {
		log.Println("WARNING: WEBHOOK_URL is empty")
	}
	return cfg, nil
}
