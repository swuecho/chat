package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type ChatSnapshotHandler struct {
	Service *svc.ChatSnapshotService
}

func NewChatSnapshotHandler(sqlc_q *sqlc_queries.Queries) *ChatSnapshotHandler {
	return &ChatSnapshotHandler{
		Service: svc.NewChatSnapshotService(sqlc_q),
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
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	uuid, err := h.Service.CreateChatSnapshot(r.Context(), chatSessionUuid, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat snapshot"))
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"uuid": uuid})
}

func (h *ChatSnapshotHandler) CreateChatBot(w http.ResponseWriter, r *http.Request) {
	chatSessionUuid := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	uuid, err := h.Service.CreateChatBot(r.Context(), chatSessionUuid, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to create chat bot").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"uuid": uuid})
}

func (h *ChatSnapshotHandler) GetChatSnapshot(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	snapshot, err := h.Service.ChatSnapshotByUUID(r.Context(), uuidStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat snapshot"))
		return
	}
	json.NewEncoder(w).Encode(snapshot)
}

func (h *ChatSnapshotHandler) ChatSnapshotMetaByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	typ := r.URL.Query().Get("type")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := int32(1)
	pageSize := int32(20)

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}
	if pageSizeStr != "" {
		if s, err := strconv.Atoi(pageSizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = int32(s)
		}
	}

	offset := (page - 1) * pageSize

	chatSnapshots, err := h.Service.ChatSnapshotMetaByUserID(r.Context(), sqlc_queries.ChatSnapshotMetaByUserIDParams{
		UserID: userID,
		Typ:    typ,
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve chat snapshots").WithDebugInfo(err.Error()))
		return
	}

	totalCount, err := h.Service.ChatSnapshotCountByUserIDAndType(r.Context(), sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams{
		UserID:  userID,
		Column2: typ,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve snapshot count").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": chatSnapshots, "page": page, "page_size": pageSize, "total": totalCount,
	})
}

func (h *ChatSnapshotHandler) UpdateChatSnapshotMetaByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if err := h.Service.UpdateChatSnapshotMetaByUUID(r.Context(), sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{
		Uuid: uuid, Title: input.Title, Summary: input.Summary, UserID: userID,
	}); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update chat snapshot metadata").WithDebugInfo(err.Error()))
		return
	}

	snapshot, err := h.Service.ChatSnapshotByUUID(r.Context(), uuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(snapshot)
}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	if err := h.Service.DeleteChatSnapshot(r.Context(), sqlc_queries.DeleteChatSnapshotParams{
		Uuid: uuid, UserID: userID,
	}); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to delete chat snapshot").WithDebugInfo(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatSnapshotHandler) ChatSnapshotSearch(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	if search == "" {
		json.NewEncoder(w).Encode([]any{})
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	chatSnapshots, err := h.Service.ChatSnapshotSearch(r.Context(), sqlc_queries.ChatSnapshotSearchParams{
		UserID: userID, Search: search,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to search chat snapshots").WithDebugInfo(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(chatSnapshots)
}
