package db

import "time"

type MessageModel struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	To           string `gorm:"size:32"`
	Content      string `gorm:"type:text"`
	Sent         bool   `gorm:"default:false;index"`
	SentAt       *time.Time
	WebhookMsgID string `gorm:"size:128"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
