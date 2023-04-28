package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

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
	router.HandleFunc("/uuid/chat_session_from_snapshot/{uuid}", h.CreateChatSessionFromSnapshot).Methods(http.MethodPost)
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
	MaxLength   int32   `json:"maxLength"`
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
	TopP        float64 `json:"topP"`
	N           int32   `json:"n"`
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

	sessionParams.MaxLength = sessionReq.MaxLength
	sessionParams.Topic = sessionReq.Topic
	sessionParams.Uuid = sessionReq.Uuid
	sessionParams.UserID = userID
	sessionParams.Temperature = sessionReq.Temperature
	sessionParams.Model = sessionReq.Model
	sessionParams.TopP = sessionReq.TopP
	sessionParams.N= sessionReq.N
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

// CreateChatSessionFromSnapshot ($uuid)
// create a new session with title of snapshot,
// create a prompt with the first message of snapshot
// create messages based on the rest of messages.
// return the new session uuid

func (h *ChatSessionHandler) CreateChatSessionFromSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshot_uuid := vars["uuid"]

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(r.Context(), snapshot_uuid)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Error retrieving chat snapshot", err)
		return
	}

	sessionTitle := snapshot.Title
	conversions := snapshot.Conversation
	var conversionsSimpleMessages []SimpleChatMessage
	json.Unmarshal(conversions, &conversionsSimpleMessages)
	promptMsg := conversionsSimpleMessages[0]
	chatPrompt, err := h.service.q.GetChatPromptByUUID(r.Context(), promptMsg.Uuid)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, eris.Wrap(err, "can not get prompt").Error(), err)
		return
	}
	originSession, err := h.service.q.GetChatSessionByUUIDWithInActive(r.Context(), chatPrompt.ChatSessionUuid)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, eris.Wrap(err, "can not get origin session").Error(), err)
		return
	}

	sessionUUID := uuid.New().String()

	session, err := h.service.q.CreateOrUpdateChatSessionByUUID(r.Context(), sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid:        sessionUUID,
		UserID:      userID,
		Topic:       sessionTitle,
		MaxLength:   originSession.MaxLength,
		Temperature: originSession.Temperature,
		Model:       originSession.Model,
		MaxTokens:   originSession.MaxTokens,
		TopP:        originSession.TopP,
		Debug:       originSession.Debug,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating chat session from snapshot: %s", err.Error())))
		return
	}

	_, err = h.service.q.CreateChatPrompt(r.Context(), sqlc_queries.CreateChatPromptParams{
		Uuid:            uuid.NewString(),
		ChatSessionUuid: sessionUUID,
		Role:            "system",
		Content:         promptMsg.Text,
		UserID:          userID,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating prompt for chat session from snapshot: %s", err.Error())))
		return
	}

	for _, message := range conversionsSimpleMessages[1:] {
		// if inversion is true, the role is user, otherwise assistant
		// Determine the role based on the inversion flag

		messageParam := sqlc_queries.CreateChatMessageParams{
			ChatSessionUuid: sessionUUID,
			Uuid:            uuid.NewString(),
			Role:            message.GetRole(),
			Content:         message.Text,
			UserID:          userID,
			Raw:             json.RawMessage([]byte("{}")),
		}
		_, err = h.service.q.CreateChatMessage(r.Context(), messageParam)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error creating messages for chat session from snapshot").Error(), err)
			return
		}

	}

	// set active session
	sessionParams := sqlc_queries.UpdateUserActiveChatSessionParams{
		UserID:          userID,
		ChatSessionUuid: session.Uuid,
	}
	_, err = h.service.q.UpdateUserActiveChatSession(r.Context(), sessionParams)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "failed to update active session").Error(), err)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"SessionUuid": session.Uuid})
}
