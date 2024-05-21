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
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	session, err := h.service.GetUserActiveChatSession(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Errorf("error: %v", err).Error(), http.StatusNotFound)
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
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var sessionParams sqlc.CreateOrUpdateUserActiveChatSessionParams
	if err := json.NewDecoder(r.Body).Decode(&sessionParams); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	// use the user_id from token
	sessionParams.UserID = userID
	session, err := h.service.CreateOrUpdateUserActiveChatSession(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, eris.Wrap(err, "fail to update or create action user session record, ").Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
