package infra_test

import (
	"testing"
	"time"

	"insider-messaging/internal/domain/entity"
	"insider-messaging/internal/infrastructure/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for testing
	// Note: Requires CGO_ENABLED=1 for sqlite driver
	// For CI/CD, consider using testcontainers or mock approach
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: SQLite requires CGO (error: %v). Run with: CGO_ENABLED=1 go test", err)
	}
	return database
}

func TestMySQLMessageRepository_Create(t *testing.T) {
	testDB := setupTestDB(t)
	repo := db.NewMySQLMessageRepository(testDB)

	msg, err := entity.NewMessage("+905551111111", "Test message", 160)
	require.NoError(t, err)

	err = repo.Create(msg)
	assert.NoError(t, err)
	
	// Verify message was created by checking unsent messages
	unsent, err := repo.GetUnsent(10)
	require.NoError(t, err)
	assert.Len(t, unsent, 1)
	assert.Equal(t, "+905551111111", unsent[0].To)
	assert.Equal(t, "Test message", unsent[0].Content)
	assert.False(t, unsent[0].Sent)
}

func TestMySQLMessageRepository_GetUnsent(t *testing.T) {
	testDB := setupTestDB(t)
	repo := db.NewMySQLMessageRepository(testDB)

	// Create unsent messages
	msg1, _ := entity.NewMessage("+905551111111", "Message 1", 160)
	msg2, _ := entity.NewMessage("+905552222222", "Message 2", 160)
	msg3, _ := entity.NewMessage("+905553333333", "Message 3", 160)
	msg3.Sent = true // This one is sent

	require.NoError(t, repo.Create(msg1))
	require.NoError(t, repo.Create(msg2))
	require.NoError(t, repo.Create(msg3))

	// Get unsent messages
	unsent, err := repo.GetUnsent(10)
	require.NoError(t, err)
	assert.Len(t, unsent, 2)
	assert.False(t, unsent[0].Sent)
	assert.False(t, unsent[1].Sent)
}

func TestMySQLMessageRepository_GetUnsent_WithLimit(t *testing.T) {
	testDB := setupTestDB(t)
	repo := db.NewMySQLMessageRepository(testDB)

	// Create multiple unsent messages
	for i := 0; i < 5; i++ {
		msg, _ := entity.NewMessage("+905551111111", "Message", 160)
		require.NoError(t, repo.Create(msg))
	}

	// Get only 2 unsent messages
	unsent, err := repo.GetUnsent(2)
	require.NoError(t, err)
	assert.Len(t, unsent, 2)
}

func TestMySQLMessageRepository_MarkSent(t *testing.T) {
	testDB := setupTestDB(t)
	repo := db.NewMySQLMessageRepository(testDB)

	msg, _ := entity.NewMessage("+905551111111", "Test message", 160)
	require.NoError(t, repo.Create(msg))

	err := repo.MarkSent(msg.ID, "webhook-123")
	assert.NoError(t, err)

	// Verify it's marked as sent
	unsent, err := repo.GetUnsent(10)
	require.NoError(t, err)
	assert.Len(t, unsent, 0)

	// Check sent messages
	sent, err := repo.ListSent()
	require.NoError(t, err)
	assert.Len(t, sent, 1)
	assert.True(t, sent[0].Sent)
	assert.Equal(t, "webhook-123", sent[0].WebhookMsgID)
	assert.NotNil(t, sent[0].SentAt)
}

func TestMySQLMessageRepository_ListSent(t *testing.T) {
	testDB := setupTestDB(t)
	repo := db.NewMySQLMessageRepository(testDB)

	// Create and send messages
	msg1, _ := entity.NewMessage("+905551111111", "Message 1", 160)
	msg2, _ := entity.NewMessage("+905552222222", "Message 2", 160)
	require.NoError(t, repo.Create(msg1))
	require.NoError(t, repo.Create(msg2))

	require.NoError(t, repo.MarkSent(msg1.ID, "webhook-1"))
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	require.NoError(t, repo.MarkSent(msg2.ID, "webhook-2"))

	sent, err := repo.ListSent()
	require.NoError(t, err)
	assert.Len(t, sent, 2)
	assert.True(t, sent[0].Sent)
	assert.True(t, sent[1].Sent)
	// Should be ordered by sent_at desc (most recent first)
	assert.Equal(t, "webhook-2", sent[0].WebhookMsgID)
}

func TestMySQLMessageRepository_Create_EmptyMessage(t *testing.T) {
	// Test entity validation, not repository
	msg, err := entity.NewMessage("", "Content", 160)
	require.Error(t, err)
	assert.Nil(t, msg)
}
