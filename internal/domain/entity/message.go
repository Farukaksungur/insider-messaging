package entity

import (
	"errors"
	"strings"
	"time"
)

// Message mesaj entity'si
// @Description Message entity with sending status
type Message struct {
	ID           uint       `json:"id" example:"1"`
	To           string     `json:"to" example:"+905551111111"`
	Content      string     `json:"content" example:"Hello, this is a test message"`
	Sent         bool       `json:"sent" example:"true"`
	SentAt       *time.Time `json:"sentAt,omitempty" example:"2024-01-01T12:00:00Z"`
	WebhookMsgID string     `json:"webhookMsgId,omitempty" example:"webhook-123"`
	CreatedAt    time.Time  `json:"createdAt" example:"2024-01-01T10:00:00Z"`
	UpdatedAt    time.Time  `json:"updatedAt" example:"2024-01-01T10:00:00Z"`
}

// NewMessage yeni bir mesaj oluşturur ve validasyon yapar
func NewMessage(to, content string, limit int) (*Message, error) {
	to = strings.TrimSpace(to)
	content = strings.TrimSpace(content)
	if to == "" || content == "" {
		return nil, errors.New("to and content required")
	}
	if len(content) > limit {
		content = content[:limit]
	}
	return &Message{To: to, Content: content}, nil
}

// MarkSent mesajı gönderilmiş olarak işaretler
func (m *Message) MarkSent(webhookId string) {
	now := time.Now().UTC()
	m.Sent = true
	m.SentAt = &now
	m.WebhookMsgID = webhookId
}
