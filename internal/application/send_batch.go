package application

import (
	"context"
	"log"
	"strconv"
	"time"

	"insider-messaging/internal/config"
	"insider-messaging/internal/domain/entity"
	"insider-messaging/internal/domain/repository"

	"github.com/go-redis/redis/v8"
)

// SenderPort mesaj gönderme işlemlerini yapan interface
type SenderPort interface {
	Send(ctx context.Context, m *entity.Message) (string, error)
}

// SendBatchUseCase mesaj gönderme işlemlerini yönetir
type SendBatchUseCase struct {
	repo   repository.MessageRepository
	sender SenderPort
	redis  *redis.Client
	cfg    *config.Config
}

// NewSendBatchUseCase yeni bir batch use case oluşturur
func NewSendBatchUseCase(r repository.MessageRepository, s SenderPort, rdb *redis.Client, cfg *config.Config) *SendBatchUseCase {
	return &SendBatchUseCase{repo: r, sender: s, redis: rdb, cfg: cfg}
}

// Execute gönderilmemiş mesajları alıp webhook'a gönderir
func (uc *SendBatchUseCase) Execute(ctx context.Context) error {
	msgs, err := uc.repo.GetUnsent(uc.cfg.MsgPerTick)
	if err != nil {
		return err
	}

	for _, m := range msgs {
		if len(m.Content) > uc.cfg.MsgCharLimit {
			m.Content = m.Content[:uc.cfg.MsgCharLimit]
		}

		msgID, err := uc.sender.Send(ctx, m)
		if err != nil {
			log.Printf("send failed id=%d err=%v", m.ID, err)
			continue
		}

		if err := uc.repo.MarkSent(m.ID, msgID); err != nil {
			log.Printf("mark sent failed id=%d err=%v", m.ID, err)
		}

		if uc.redis != nil {
			key := "message:" + strconv.FormatUint(uint64(m.ID), 10)
			now := time.Now().UTC().Format(time.RFC3339)
			uc.redis.HSet(ctx, key, map[string]interface{}{
				"webhook_id": msgID,
				"sent_at":    now,
			})
		}
	}
	return nil
}
