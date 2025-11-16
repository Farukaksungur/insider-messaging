package repository

import "insider-messaging/internal/domain/entity"

type MessageRepository interface {
	GetUnsent(limit int) ([]*entity.Message, error)
	MarkSent(id uint, webhookMsgId string) error
	ListSent() ([]*entity.Message, error)
	Create(msg *entity.Message) error
}
