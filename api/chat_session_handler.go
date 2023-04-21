package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
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
		http.Error(w, eris.Wrap(err, "invalid chat session ID ").Error(), http.StatusBadRequest)
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
	session_resp := &ChatSessionResponse{}
	if err != nil {
		if err == sql.ErrNoRows {
			session_resp.Uuid = session.Uuid
			session_resp.MaxLength = 10
			json.NewEncoder(w).Encode(session_resp)
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}
	session_resp.Uuid = session.Uuid
	session_resp.Topic = session.Topic
	session_resp.MaxLength = session.MaxLength
	session_resp.CreatedAt = session.CreatedAt
	session_resp.UpdatedAt = session.UpdatedAt
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
	userIDInt, err := getUserID(ctx)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sessionParams.UserID = userIDInt
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
		http.Error(w, eris.Wrap(err, "fail to update or create action user session record, ").Error(), http.StatusInternalServerError)
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
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sessionParams.UserID = userID
	session, err := h.service.UpdateChatSessionByUUID(r.Context(), sessionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

type UpdateChatSessionRequest struct {
	Uuid        string  `json:"uuid"`
	Topic       string  `json:"topic"`
	KeepLength  int32   `json:"keepLength"`
	MaxLength   int32   `json:"maxLength"`
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
	TopP        float64 `json:"topP"`
	MaxTokens   int32   `json:"maxTokens"`
	Debug       bool    `json:"debug"`
}

// UpdateChatSessionByUUID updates a chat session by its UUID
func (h *ChatSessionHandler) CreateOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
	var sessionReq UpdateChatSessionRequest
	err := json.NewDecoder(r.Body).Decode(&sessionReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = 10
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	var sessionParams sqlc_queries.CreateOrUpdateChatSessionByUUIDParams

	sessionParams.KeepLength = sessionReq.KeepLength
	sessionParams.MaxLength = sessionReq.MaxLength
	sessionParams.Topic = sessionReq.Topic
	sessionParams.Uuid = sessionReq.Uuid
	sessionParams.UserID = userID
	sessionParams.Temperature = sessionReq.Temperature
	sessionParams.Model = sessionReq.Model
	sessionParams.TopP = sessionReq.TopP
	sessionParams.MaxTokens = sessionReq.MaxTokens
	sessionParams.Debug = sessionReq.Debug
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
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sessionParams.UserID = userID

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
