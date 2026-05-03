package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type ChatCommentHandler struct {
	service *svc.ChatCommentService
}

func NewChatCommentHandler(sqlc_q *sqlc_queries.Queries) *ChatCommentHandler {
	return &ChatCommentHandler{
		service: svc.NewChatCommentService(sqlc_q),
	}
}

func (h *ChatCommentHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/chat_sessions/{sessionUUID}/chat_messages/{messageUUID}/comments", h.CreateChatComment).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_sessions/{sessionUUID}/comments", h.GetCommentsBySessionUUID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/{messageUUID}/comments", h.GetCommentsByMessageUUID).Methods(http.MethodGet)
}

func (h *ChatCommentHandler) CreateChatComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionUUID := vars["sessionUUID"]
	messageUUID := vars["messageUUID"]

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("unauthorized").WithDebugInfo(err.Error()))
		return
	}

	comment, err := h.service.CreateChatComment(r.Context(), sqlc_queries.CreateChatCommentParams{
		Uuid:            uuid.New().String(),
		ChatSessionUuid: sessionUUID,
		ChatMessageUuid: messageUUID,
		Content:         req.Content,
		CreatedBy:       userID,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat comment"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *ChatCommentHandler) GetCommentsBySessionUUID(w http.ResponseWriter, r *http.Request) {
	sessionUUID := mux.Vars(r)["sessionUUID"]

	comments, err := h.service.GetCommentsBySessionUUID(r.Context(), sessionUUID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get comments by session"))
		return
	}

	json.NewEncoder(w).Encode(comments)
}

func (h *ChatCommentHandler) GetCommentsByMessageUUID(w http.ResponseWriter, r *http.Request) {
	messageUUID := mux.Vars(r)["messageUUID"]

	comments, err := h.service.GetCommentsByMessageUUID(r.Context(), messageUUID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get comments by message"))
		return
	}

	json.NewEncoder(w).Encode(comments)
}
