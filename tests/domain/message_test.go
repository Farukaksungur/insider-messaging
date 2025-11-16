package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"insider-messaging/internal/domain/entity"
)

func TestNewMessage_Truncate(t *testing.T) {
	m, err := entity.NewMessage("+90", "hello world", 5)
	assert.NoError(t, err)
	assert.Equal(t, "hello", m.Content)
}
