package db

import (
	"insider-messaging/internal/domain/entity"
	"insider-messaging/internal/domain/repository"

	"gorm.io/gorm"
)

type MySQLMessageRepository struct {
	db *gorm.DB
}

// NewMySQLMessageRepository yeni bir MySQL repository oluşturur ve tabloyu hazırlar
func NewMySQLMessageRepository(db *gorm.DB) repository.MessageRepository {
	db.AutoMigrate(&MessageModel{})
	return &MySQLMessageRepository{db: db}
}

// Create yeni bir mesaj kaydı oluşturur
func (r *MySQLMessageRepository) Create(msg *entity.Message) error {
	row := MessageModel{To: msg.To, Content: msg.Content, Sent: false}
	return r.db.Create(&row).Error
}

// GetUnsent gönderilmemiş mesajları getirir, limit kadar
func (r *MySQLMessageRepository) GetUnsent(limit int) ([]*entity.Message, error) {
	var rows []MessageModel
	if err := r.db.Where("sent = ?", false).Order("created_at asc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	msgs := make([]*entity.Message, 0, len(rows))
	for _, rr := range rows {
		msgs = append(msgs, &entity.Message{
			ID: rr.ID, To: rr.To, Content: rr.Content, Sent: rr.Sent,
			SentAt: rr.SentAt, WebhookMsgID: rr.WebhookMsgID,
			CreatedAt: rr.CreatedAt, UpdatedAt: rr.UpdatedAt,
		})
	}
	return msgs, nil
}

// MarkSent mesajı gönderilmiş olarak işaretler
func (r *MySQLMessageRepository) MarkSent(id uint, webhookMsgId string) error {
	return r.db.Model(&MessageModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"sent": true, "webhook_msg_id": webhookMsgId, "sent_at": gorm.Expr("NOW()"),
	}).Error
}

// ListSent gönderilmiş tüm mesajları getirir
func (r *MySQLMessageRepository) ListSent() ([]*entity.Message, error) {
	var rows []MessageModel
	if err := r.db.Where("sent = ?", true).Order("sent_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	msgs := make([]*entity.Message, 0, len(rows))
	for _, rr := range rows {
		msgs = append(msgs, &entity.Message{
			ID: rr.ID, To: rr.To, Content: rr.Content, Sent: rr.Sent,
			SentAt: rr.SentAt, WebhookMsgID: rr.WebhookMsgID,
			CreatedAt: rr.CreatedAt, UpdatedAt: rr.UpdatedAt,
		})
	}
	return msgs, nil
}
