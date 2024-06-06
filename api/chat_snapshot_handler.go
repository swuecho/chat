package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatSnapshotHandler struct {
	service *ChatSnapshotService
}

func NewChatSnapshotHandler(sqlc_q *sqlc_queries.Queries) *ChatSnapshotHandler {
	return &ChatSnapshotHandler{
		service: NewChatSnapshotService(sqlc_q),
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

func (h *ChatSnapshotHandler) CreateChatSnapshot(w http.ResponseWriter, r *http.Request) {
	chatSessionUuid := mux.Vars(r)["uuid"]
	user_id, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	uuid, err := h.service.CreateChatSnapshot(r.Context(), chatSessionUuid, user_id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"uuid": uuid,
		})

}

func (h *ChatSnapshotHandler) GetChatSnapshot(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	snapshot, err := h.service.q.ChatSnapshotByUUID(r.Context(), uuidStr)
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
	chatSnapshots, err := h.service.q.ChatSnapshotMetaByUserID(r.Context(), userID)

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
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	err = h.service.q.UpdateChatSnapshotMetaByUUID(r.Context(), sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{
		Uuid:    uuid,
		Title:   input.Title,
		Summary: input.Summary,
		UserID:  userID,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(r.Context(), uuid)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	json.NewEncoder(w).Encode(snapshot)

}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusForbidden, err.Error(), err)
		return
	}

	_, err = h.service.q.DeleteChatSnapshot(r.Context(), sqlc_queries.DeleteChatSnapshotParams{
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

	chatSnapshots, err := h.service.q.ChatSnapshotSearch(r.Context(), sqlc_queries.ChatSnapshotSearchParams{
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
