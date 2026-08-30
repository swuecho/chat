package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatCommentHandler struct {
	service *svc.ChatCommentService
}

func NewChatCommentHandler(service *svc.ChatCommentService) *ChatCommentHandler {
	return &ChatCommentHandler{service: service}
}

func (h *ChatCommentHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/chat_sessions/{sessionUUID}/chat_messages/{messageUUID}/comments", endpoint(h.CreateChatComment)).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_sessions/{sessionUUID}/comments", endpoint(h.GetCommentsBySessionUUID)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/{messageUUID}/comments", endpoint(h.GetCommentsByMessageUUID)).Methods(http.MethodGet)
}

func (h *ChatCommentHandler) CreateChatComment(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	sessionUUID := vars["sessionUUID"]
	messageUUID := vars["messageUUID"]

	var req createCommentRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	comment, err := h.service.CreateChatComment(r.Context(), svc.CreateChatCommentInput{
		Uuid:            uuid.New().String(),
		ChatSessionUuid: sessionUUID,
		ChatMessageUuid: messageUUID,
		Content:         req.Content,
		CreatedBy:       userID,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat comment")
	}
	return respondJSON(w, http.StatusCreated, comment)
}

func (h *ChatCommentHandler) GetCommentsBySessionUUID(w http.ResponseWriter, r *http.Request) error {
	sessionUUID := mux.Vars(r)["sessionUUID"]

	comments, err := h.service.GetCommentsBySessionUUID(r.Context(), sessionUUID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get comments by session")
	}
	return respondJSON(w, http.StatusOK, comments)
}

func (h *ChatCommentHandler) GetCommentsByMessageUUID(w http.ResponseWriter, r *http.Request) error {
	messageUUID := mux.Vars(r)["messageUUID"]

	comments, err := h.service.GetCommentsByMessageUUID(r.Context(), messageUUID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get comments by message")
	}
	return respondJSON(w, http.StatusOK, comments)
}
