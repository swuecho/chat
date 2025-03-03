package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatCommentHandler struct {
	service *ChatCommentService
}

func NewChatCommentHandler(sqlc_q *sqlc_queries.Queries) *ChatCommentHandler {
	chatCommentService := NewChatCommentService(sqlc_q)
	return &ChatCommentHandler{
		service: chatCommentService,
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *ChatCommentHandler) GetCommentsBySessionUUID(w http.ResponseWriter, r *http.Request) {
	sessionUUID := mux.Vars(r)["sessionUUID"]

	comments, err := h.service.GetCommentsBySessionUUID(r.Context(), sessionUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

func (h *ChatCommentHandler) GetCommentsByMessageUUID(w http.ResponseWriter, r *http.Request) {
	messageUUID := mux.Vars(r)["messageUUID"]

	comments, err := h.service.GetCommentsByMessageUUID(r.Context(), messageUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}
