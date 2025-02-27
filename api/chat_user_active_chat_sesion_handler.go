package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rotisserie/eris"
	sqlc "github.com/swuecho/chat_backend/sqlc_queries"
)

type UserActiveChatSessionHandler struct {
	service *UserActiveChatSessionService
}

func NewUserActiveChatSessionHandler(sqlc_q *sqlc.Queries) *UserActiveChatSessionHandler {
	activeSessionService := NewUserActiveChatSessionService(sqlc_q)

	return &UserActiveChatSessionHandler{
		service: activeSessionService,
	}
}

func (h *UserActiveChatSessionHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/user_active_chat_session", h.GetUserActiveChatSessionHandler).Methods(http.MethodGet)
	router.HandleFunc("/uuid/user_active_chat_session", h.CreateOrUpdateUserActiveChatSessionHandler).Methods(http.MethodPut)
}

// GetUserActiveChatSessionHandler handles GET requests to get a session by user_id.
func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	session, err := h.service.GetUserActiveChatSession(r.Context(), userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get active chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// CreateOrUpdateUserActiveChatSessionHandler handles POST requests to create a new session.
func (h *UserActiveChatSessionHandler) CreateOrUpdateUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	var sessionParams sqlc.CreateOrUpdateUserActiveChatSessionParams
	if err := json.NewDecoder(r.Body).Decode(&sessionParams); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse request body"))
		return
	}
	// use the user_id from token
	sessionParams.UserID = userID
	session, err := h.service.CreateOrUpdateUserActiveChatSession(r.Context(), sessionParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to create or update active chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
