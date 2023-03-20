package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chatgpt_backend/sqlc_queries"
)

type ChatSessionHandler struct {
	service *ChatSessionService
}

func NewChatSessionHandler(service *ChatSessionService) *ChatSessionHandler {
	return &ChatSessionHandler{
		service: service,
	}
}

func (h *ChatSessionHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_sessions", h.CreateChatSession).Methods(http.MethodPost)

	router.HandleFunc("/chat_sessions/users", h.GetSimpleChatSessionsByUserID).Methods(http.MethodGet)
	router.HandleFunc("/chat_sessions/{id}", h.GetChatSessionByID).Methods(http.MethodGet)
	router.HandleFunc("/chat_sessions/{id}", h.UpdateChatSession).Methods(http.MethodPut)
	router.HandleFunc("/chat_sessions/{id}", h.DeleteChatSession).Methods(http.MethodDelete)
	router.HandleFunc("/chat_sessions", h.GetAllChatSessions).Methods(http.MethodGet)

	router.HandleFunc("/uuid/chat_sessions/max_length/{uuid}", h.UpdateSessionMaxLength).Methods("PUT")
	router.HandleFunc("/uuid/chat_sessions/topic/{uuid}", h.UpdateChatSessionTopicByUUID).Methods("PUT")
	router.HandleFunc("/uuid/chat_sessions/{uuid}", h.GetChatSessionByUUID).Methods("GET")
	router.HandleFunc("/uuid/chat_sessions/{uuid}", h.CreateOrUpdateChatSessionByUUID).Methods("PUT")
	router.HandleFunc("/uuid/chat_sessions/{uuid}", h.DeleteChatSessionByUUID).Methods("DELETE")
	router.HandleFunc("/uuid/chat_sessions", h.CreateChatSessionByUUID).Methods("POST")
}

func (h *ChatSessionHandler) CreateChatSession(w http.ResponseWriter, r *http.Request) {
	var sessionParams sqlc_queries.CreateChatSessionParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, err := h.service.CreateChatSession(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *ChatSessionHandler) GetChatSessionByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Errorf("invalid chat session ID %w", err).Error(), http.StatusBadRequest)
		return
	}
	session, err := h.service.GetChatSessionByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *ChatSessionHandler) UpdateChatSession(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat session ID", http.StatusBadRequest)
		return
	}
	var sessionParams sqlc_queries.UpdateChatSessionParams
	err = json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionParams.ID = int32(id)
	session, err := h.service.UpdateChatSession(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *ChatSessionHandler) DeleteChatSession(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat session ID", http.StatusBadRequest)
		return
	}
	err = h.service.DeleteChatSession(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatSessionHandler) GetAllChatSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.service.GetAllChatSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *ChatSessionHandler) GetChatSessionsByUserID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	sessions, err := h.service.GetChatSessionsByUserID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

// GetChatSessionByUUID returns a chat session by its UUID
func (h *ChatSessionHandler) GetChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	session, err := h.service.GetChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	session_resp := &ChatSessionResponse{
		Uuid:      session.Uuid,
		Topic:     session.Topic,
		MaxLength: session.MaxLength,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
	json.NewEncoder(w).Encode(session_resp)
}

// CreateChatSessionByUUID creates a chat session by its UUID
func (h *ChatSessionHandler) CreateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var sessionParams sqlc_queries.CreateChatSessionParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", http.StatusBadRequest)
		return
	}
	sessionParams.UserID = int32(userIDInt)
	sessionParams.MaxLength = 10
	session, err := h.service.CreateChatSession(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// set active chat session when creating a new chat session
	sessionParams.UserID = int32(userIDInt)
	_, err = h.service.q.CreateOrUpdateUserActiveChatSession(r.Context(),
		sqlc_queries.CreateOrUpdateUserActiveChatSessionParams{
			UserID:          session.UserID,
			ChatSessionUuid: session.Uuid,
		})
	if err != nil {
		http.Error(w, fmt.Errorf("fail to update or create action user session record, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// UpdateChatSessionByUUID updates a chat session by its UUID
func (h *ChatSessionHandler) UpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var sessionParams sqlc_queries.UpdateChatSessionByUUIDParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionParams.Uuid = uuid

	ctx := r.Context()
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", http.StatusBadRequest)
		return
	}

	sessionParams.UserID = int32(userIDInt)
	session, err := h.service.UpdateChatSessionByUUID(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// UpdateChatSessionByUUID updates a chat session by its UUID
func (h *ChatSessionHandler) CreateOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var sessionParams sqlc_queries.CreateOrUpdateChatSessionByUUIDParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionParams.Uuid = uuid

	ctx := r.Context()
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", http.StatusBadRequest)
		return
	}

	sessionParams.UserID = int32(userIDInt)
	session, err := h.service.CreateOrUpdateChatSessionByUUID(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// DeleteChatSessionByUUID deletes a chat session by its UUID
func (h *ChatSessionHandler) DeleteChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatSessionByUUID(r.Context(), uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetSimpleChatSessionsByUserID returns a list of simple chat sessions by user ID
func (h *ChatSessionHandler) GetSimpleChatSessionsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := ctx.Value(userContextKey).(string)
	print("xx", idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	sessions, err := h.service.GetSimpleChatSessionsByUserID(ctx, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

// UpdateChatSessionTopicByUUID updates a chat session topic by its UUID
func (h *ChatSessionHandler) UpdateChatSessionTopicByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var sessionParams sqlc_queries.UpdateChatSessionTopicByUUIDParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionParams.Uuid = uuid

	ctx := r.Context()
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", http.StatusBadRequest)
		return
	}

	sessionParams.UserID = int32(userIDInt)

	session, err := h.service.UpdateChatSessionTopicByUUID(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

// UpdateSessionMaxLength
func (h *ChatSessionHandler) UpdateSessionMaxLength(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var sessionParams sqlc_queries.UpdateSessionMaxLengthParams
	err := json.NewDecoder(r.Body).Decode(&sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionParams.Uuid = uuid

	session, err := h.service.UpdateSessionMaxLength(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}
