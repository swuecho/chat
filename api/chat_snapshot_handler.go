package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatSnapshotHandler handles requests related to chat snapshots
type ChatSnapshotHandler struct {
	service *ChatSnapshotService
}

// NewChatSnapshotHandler creates a new handler instance
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
	router.HandleFunc("/uuid/chat_bot/{uuid}", h.CreateChatBot).Methods(http.MethodPost)
}

func (h *ChatSnapshotHandler) CreateChatSnapshot(w http.ResponseWriter, r *http.Request) {
	chatSessionUuid := mux.Vars(r)["uuid"]
	user_id, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	uuid, err := h.service.CreateChatSnapshot(r.Context(), chatSessionUuid, user_id)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to create chat snapshot")
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"uuid": uuid,
		})

}

func (h *ChatSnapshotHandler) CreateChatBot(w http.ResponseWriter, r *http.Request) {
	chatSessionUuid := mux.Vars(r)["uuid"]
	user_id, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	uuid, err := h.service.CreateChatBot(r.Context(), chatSessionUuid, user_id)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat bot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
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
		apiErr := WrapError(MapDatabaseError(err), "Failed to get chat snapshot")
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(snapshot)

}

func (h *ChatSnapshotHandler) ChatSnapshotMetaByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	// get type from query
	typ := r.URL.Query().Get("type")

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := int32(1)    // Default to page 1
	pageSize := int32(20) // Default to 20 items per page

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = int32(parsedPage)
		}
	}

	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err == nil && parsedPageSize > 0 && parsedPageSize <= 100 {
			pageSize = int32(parsedPageSize)
		}
	}

	offset := (page - 1) * pageSize

	chatSnapshots, err := h.service.q.ChatSnapshotMetaByUserID(r.Context(), sqlc_queries.ChatSnapshotMetaByUserIDParams{
		UserID: userID,
		Typ:    typ,
		Limit:  pageSize,
		Offset: offset,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve chat snapshots"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Get total count for pagination
	totalCount, err := h.service.q.ChatSnapshotCountByUserIDAndType(r.Context(), sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams{
		UserID: userID,
		Column2: typ,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve snapshot count"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       chatSnapshots,
		"page":       page,
		"page_size":  pageSize,
		"total":      totalCount,
	})
}
func (h *ChatSnapshotHandler) UpdateChatSnapshotMetaByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	err = h.service.q.UpdateChatSnapshotMetaByUUID(r.Context(), sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{
		Uuid:    uuid,
		Title:   input.Title,
		Summary: input.Summary,
		UserID:  userID,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat snapshot metadata"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(r.Context(), uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	json.NewEncoder(w).Encode(snapshot)

}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	_, err = h.service.q.DeleteChatSnapshot(r.Context(), sqlc_queries.DeleteChatSnapshotParams{
		Uuid:   uuid,
		UserID: userID,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat snapshot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
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
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	chatSnapshots, err := h.service.q.ChatSnapshotSearch(r.Context(), sqlc_queries.ChatSnapshotSearchParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to search chat snapshots"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatSnapshots)
}
