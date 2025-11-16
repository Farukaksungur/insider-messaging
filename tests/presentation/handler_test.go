package presentation_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"insider-messaging/internal/config"
	"insider-messaging/internal/domain/entity"
	"insider-messaging/internal/presentation/api"

	"github.com/stretchr/testify/assert"
)

/*
	------------------------------
	  MOCK SCHEDULER

--------------------------------
*/
type mockScheduler struct {
	startCalled bool
	stopCalled  bool
	running     bool
}

func (m *mockScheduler) Start() { m.startCalled = true; m.running = true }
func (m *mockScheduler) Stop()  { m.stopCalled = true; m.running = false }
func (m *mockScheduler) IsRunning() bool {
	return m.running
}

/*
	------------------------------
	  MOCK REPOSITORY

--------------------------------
*/
type mockRepo struct {
	createCalled bool
	sentCalled   bool
	sentList     []*entity.Message
	createErr    error
	listErr      error
}

func (m *mockRepo) Create(msg *entity.Message) error {
	m.createCalled = true
	return m.createErr
}

func (m *mockRepo) GetUnsent(limit int) ([]*entity.Message, error) {
	return nil, nil
}

func (m *mockRepo) MarkSent(id uint, wid string) error {
	return nil
}

func (m *mockRepo) ListSent() ([]*entity.Message, error) {
	m.sentCalled = true
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.sentList, nil
}

/* ------------------------------
     TESTS
--------------------------------*/

func getTestConfig() *config.Config {
	return &config.Config{
		MsgCharLimit:          160,
		WebhookTimeoutSeconds: 30,
		// API key not set for tests (development mode)
	}
}

func Test_StartStop_Start(t *testing.T) {
	mSched := &mockScheduler{}
	mRepo := &mockRepo{}

	h := api.NewHandler(mSched, mRepo, getTestConfig())

	req := httptest.NewRequest("GET", "/api/auto?action=start", nil)
	w := httptest.NewRecorder()

	h.StartStop(w, req)

	assert.True(t, mSched.startCalled)
	assert.Equal(t, 200, w.Result().StatusCode)
}

func Test_StartStop_Stop(t *testing.T) {
	mSched := &mockScheduler{}
	mRepo := &mockRepo{}

	h := api.NewHandler(mSched, mRepo, getTestConfig())

	req := httptest.NewRequest("GET", "/api/auto?action=stop", nil)
	w := httptest.NewRecorder()

	h.StartStop(w, req)

	assert.True(t, mSched.stopCalled)
	assert.Equal(t, 200, w.Result().StatusCode)
}

func Test_ListSent_Success(t *testing.T) {
	mSched := &mockScheduler{}
	mRepo := &mockRepo{
		sentList: []*entity.Message{
			{ID: 1, To: "+90555", Content: "hi", Sent: true},
		},
	}

	h := api.NewHandler(mSched, mRepo, getTestConfig())

	req := httptest.NewRequest("GET", "/api/sent", nil)
	w := httptest.NewRecorder()

	h.ListSent(w, req)

	assert.True(t, mRepo.sentCalled)
	assert.Equal(t, 200, w.Result().StatusCode)

	var out []*entity.Message
	err := json.NewDecoder(w.Body).Decode(&out)
	assert.NoError(t, err)
	assert.Len(t, out, 1)
	assert.Equal(t, uint(1), out[0].ID)
}

func Test_CreateMessage(t *testing.T) {
	mSched := &mockScheduler{}
	mRepo := &mockRepo{}

	body := bytes.NewBuffer([]byte(`{"to":"+905551111111","content":"hello"}`))
	req := httptest.NewRequest("POST", "/api/messages", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h := api.NewHandler(mSched, mRepo, getTestConfig())
	h.CreateMessage(w, req)

	assert.True(t, mRepo.createCalled)
	assert.Equal(t, 201, w.Code)
}
