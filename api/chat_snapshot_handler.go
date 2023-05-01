package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatSnapshotHandler struct {
	chatMessageService *ChatMessageService
}

func NewChatSnapshotHandler(service *ChatMessageService) *ChatSnapshotHandler {
	return &ChatSnapshotHandler{
		chatMessageService: service,
	}
}

func (h *ChatSnapshotHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/chat_snapshot/all", h.ChatSnapshotMetaByUserID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.GetChatSnapshot).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.CreateChatSnapshot).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.UpdateChatSnapshotMetaByUUID).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.DeleteChatSnapshot).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_snapshot_search", h.ChatSnapshotSearch).Methods(http.MethodGet)
}

// save all chat messages to database

func (h *ChatSnapshotHandler) CreateChatSnapshot(w http.ResponseWriter, r *http.Request) {
	chatSessionUuid := mux.Vars(r)["uuid"]
	user_id, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	chatSession, err := h.chatMessageService.q.GetChatSessionByUUID(r.Context(), chatSessionUuid)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	// TODO: fix hardcode
	simple_msgs, err := h.chatMessageService.GetChatHistoryBySessionUUID(r.Context(), chatSessionUuid, 1, 10000)
	text := lo.Reduce(simple_msgs, func(acc string, curr SimpleChatMessage, _ int) string {
		return acc + curr.Text
	},
		"")
	// save all simple_msgs to a jsonb field in chat_snapshot
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// simple_msgs to RawMessage
	simple_msgs_raw, err := json.Marshal(simple_msgs)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	snapshot_uuid := uuid.New().String()
	chatSessionMessage, err := json.Marshal(chatSession)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}
	one, err := h.chatMessageService.q.CreateChatSnapshot(r.Context(), sqlc_queries.CreateChatSnapshotParams{
		Uuid:         snapshot_uuid,
		Model:        chatSession.Model,
		Title:        firstN(chatSession.Topic, 100),
		UserID:       user_id,
		Session:      chatSessionMessage,
		Tags:         json.RawMessage([]byte("{}")),
		Text:         text,
		Conversation: simple_msgs_raw,
	})
	if err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"uuid": one.Uuid,
		})

}

func (h *ChatSnapshotHandler) GetChatSnapshot(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	snapshot, err := h.chatMessageService.q.ChatSnapshotByUUID(r.Context(), uuidStr)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	json.NewEncoder(w).Encode(snapshot)

}

func (h *ChatSnapshotHandler) ChatSnapshotMetaByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}
	chatSnapshots, err := h.chatMessageService.q.ChatSnapshotMetaByUserID(r.Context(), userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatSnapshots)
}
func (h *ChatSnapshotHandler) UpdateChatSnapshotMetaByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request body"))
		return
	}
	log.Println(input)

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	err = h.chatMessageService.q.UpdateChatSnapshotMetaByUUID(r.Context(), sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{
		Uuid:    uuid,
		Title:   input.Title,
		Summary: input.Summary,
		UserID:  userID,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusForbidden, err.Error(), err)
		return
	}

	_, err = h.chatMessageService.q.DeleteChatSnapshot(r.Context(), sqlc_queries.DeleteChatSnapshotParams{
		Uuid:   uuid,
		UserID: userID,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

}

func (h *ChatSnapshotHandler) ChatSnapshotSearch(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	if search == "" {
		w.WriteHeader(http.StatusOK)
		var emptySlice []any // create an empty slice of integers
		json.NewEncoder(w).Encode(emptySlice)
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	chatSnapshots, err := h.chatMessageService.q.ChatSnapshotSearch(r.Context(), sqlc_queries.ChatSnapshotSearchParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "failed to retrieve chat snapshots").Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatSnapshots)
}
