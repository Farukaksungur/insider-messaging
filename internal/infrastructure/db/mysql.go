package db

import (
	"fmt"
	"log"
	"time"

	"insider-messaging/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMySQL MySQL veritabanı bağlantısı oluşturur, bağlantı başarısız olursa tekrar dener
func NewMySQL(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	var db *gorm.DB
	var err error
	maxRetries := 10
	retryDelay := 3 * time.Second
	connected := false

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					connected = true
					break
				}
			}
		}

		if i < maxRetries-1 {
			log.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if !connected {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	log.Println("Connected to MySQL")
	return db, nil
}
