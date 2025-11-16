package sender

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"insider-messaging/internal/config"
	"insider-messaging/internal/domain/entity"
)

type WebhookSender struct {
	cfg    *config.Config
	client *http.Client
}

// NewWebhookSender yeni bir webhook sender oluşturur
func NewWebhookSender(cfg *config.Config) *WebhookSender {
	timeout := 10 * time.Second
	if cfg.WebhookTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.WebhookTimeoutSeconds) * time.Second
	}
	return &WebhookSender{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

type webhookReq struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type webhookResp struct {
	Message   string `json:"message"`
	MessageId string `json:"messageId"`
}

// Send mesajı webhook URL'ine gönderir ve dönen messageId'yi alır
func (s *WebhookSender) Send(ctx context.Context, m *entity.Message) (string, error) {
	payload := webhookReq{To: m.To, Content: m.Content}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.WebhookURL, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.WebhookAuthKey != "" {
		req.Header.Set("x-ins-auth-key", s.cfg.WebhookAuthKey)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	bodyBytes := make([]byte, 0, 512)
	buf := make([]byte, 512)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	var wr webhookResp
	if err := json.Unmarshal(bodyBytes, &wr); err != nil {
		responseStr := string(bodyBytes)
		if len(responseStr) > 200 {
			responseStr = responseStr[:200] + "..."
		}
		return "", fmt.Errorf("failed to decode response (status %d): %v. Response body: %s", resp.StatusCode, err, responseStr)
	}

	if wr.MessageId == "" || wr.MessageId == "{{uuid}}" {
		uuid, err := generateUUID()
		if err != nil {
			return "", fmt.Errorf("failed to generate UUID: %w", err)
		}
		return uuid, nil
	}

	return wr.MessageId, nil
}

// generateUUID rastgele bir UUID v4 oluşturur
func generateUUID() (string, error) {
	uuidBytes := make([]byte, 16)
	if _, err := rand.Read(uuidBytes); err != nil {
		return "", err
	}

	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x40
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80

	uuid := fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hex.EncodeToString(uuidBytes[0:4]),
		hex.EncodeToString(uuidBytes[4:6]),
		hex.EncodeToString(uuidBytes[6:8]),
		hex.EncodeToString(uuidBytes[8:10]),
		hex.EncodeToString(uuidBytes[10:16]),
	)

	return uuid, nil
}
