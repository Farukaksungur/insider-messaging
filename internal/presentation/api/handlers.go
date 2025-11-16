package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"insider-messaging/internal/application"
	"insider-messaging/internal/config"
	"insider-messaging/internal/domain/entity"
	"insider-messaging/internal/domain/repository"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// logError hatayı loglar ve HTTP response döner
func logError(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("API error: %s", message)
	http.Error(w, message, statusCode)
}

type StatusResponse struct {
	Status string `json:"status" example:"started"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"use ?action=start or ?action=stop"`
	Message string `json:"message,omitempty" example:"Detailed error message"`
	Code    string `json:"code,omitempty" example:"INVALID_ACTION"`
}

type CreateMessageRequest struct {
	To      string `json:"to" example:"+905551111111" binding:"required"`
	Content string `json:"content" example:"Hello, this is a test message" binding:"required"`
}

type Handler struct {
	sched application.SchedulerController
	repo  repository.MessageRepository
	cfg   *config.Config
}

// NewHandler yeni bir handler oluşturur
func NewHandler(s application.SchedulerController, r repository.MessageRepository, cfg *config.Config) *Handler {
	return &Handler{sched: s, repo: r, cfg: cfg}
}

// StartStop scheduler'ı başlatır veya durdurur
// @Summary      Start or stop automatic message sending
// @Description  Control the automatic message sending scheduler
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string  true  "API Key for authentication"
// @Param        action     query     string  true  "Action to perform"  Enums(start, stop)
// @Success      200        {object}  StatusResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      401        {object}  ErrorResponse
// @Router       /auto [post]
// @Router       /auto [get]
func (h *Handler) StartStop(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")

	switch action {
	case "start":
		h.sched.Start()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(StatusResponse{Status: "started"}); err != nil {
			logError(w, "failed to encode response", http.StatusInternalServerError)
		}
		return
	case "stop":
		h.sched.Stop()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(StatusResponse{Status: "stopped"}); err != nil {
			logError(w, "failed to encode response", http.StatusInternalServerError)
		}
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid action parameter",
			Message: "Action must be either 'start' or 'stop'",
			Code:    "INVALID_ACTION",
		})
		return
	}
}

// ListSent gönderilmiş tüm mesajları listeler
// @Summary      List all sent messages
// @Description  Retrieve a list of all messages that have been successfully sent
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string  true  "API Key for authentication"
// @Success      200        {array}   entity.Message
// @Failure      401        {object}  ErrorResponse
// @Failure      500        {object}  ErrorResponse
// @Router       /sent [get]
func (h *Handler) ListSent(w http.ResponseWriter, r *http.Request) {
	msgs, err := h.repo.ListSent()
	if err != nil {
		logError(w, "Failed to retrieve sent messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		logError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateMessage yeni bir mesaj oluşturur
// @Summary      Create a new message
// @Description  Create a new message that will be sent automatically in the next batch
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string                true  "API Key for authentication"
// @Param        message    body      CreateMessageRequest  true  "Message data"
// @Success      201        {object}  entity.Message
// @Failure      400        {object}  ErrorResponse
// @Failure      401        {object}  ErrorResponse
// @Failure      500        {object}  ErrorResponse
// @Router       /messages [post]
func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var in CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid request payload",
			Message: "Request body must be valid JSON",
			Code:    "INVALID_PAYLOAD",
		})
		return
	}

	if !phoneRegex.MatchString(in.To) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid phone number format",
			Message: "Phone number must be in international format (e.g., +905551111111)",
			Code:    "INVALID_PHONE_NUMBER",
		})
		return
	}

	if in.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Content cannot be empty",
			Message: "Message content is required",
			Code:    "EMPTY_CONTENT",
		})
		return
	}

	msg, err := entity.NewMessage(in.To, in.Content, h.cfg.MsgCharLimit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	if err := h.repo.Create(msg); err != nil {
		logError(w, "Failed to create message in database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		logError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
