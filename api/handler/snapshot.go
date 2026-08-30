package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatSnapshotHandler struct {
	Service *svc.ChatSnapshotService
}

func NewChatSnapshotHandler(service *svc.ChatSnapshotService) *ChatSnapshotHandler {
	return &ChatSnapshotHandler{Service: service}
}

func (h *ChatSnapshotHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/chat_snapshot/all", h.ChatSnapshotMetaByUserID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.GetChatSnapshot).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.CreateChatSnapshot).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.UpdateChatSnapshotMetaByUUID).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", h.DeleteChatSnapshot).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_snapshot_search", h.ChatSnapshotSearch).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_bot/{uuid}", h.CreateChatBot).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_bot/{uuid}/model", h.UpdateChatBotModel).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_bot/{uuid}/settings", h.UpdateChatBotSettings).Methods(http.MethodPut)
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
	json.NewEncoder(w).Encode(uuidHTTPResponse{UUID: uuid})
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
	json.NewEncoder(w).Encode(uuidHTTPResponse{UUID: uuid})
}

func (h *ChatSnapshotHandler) UpdateChatBotSettings(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
		Model   string `json:"model"`
	}
	if err := DecodeJSON(r, &input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Model = strings.TrimSpace(input.Model)
	if input.Title == "" || input.Model == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Title and model are required"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	snapshot, err := h.Service.UpdateChatBotSettings(r.Context(), svc.UpdateChatBotSettingsCommand{UUID: uuid, UserID: userID, Title: input.Title, Summary: input.Summary, Model: input.Model})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Bot or enabled model").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) UpdateChatBotModel(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var input struct {
		Model string `json:"model"`
	}
	if err := DecodeJSON(r, &input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}
	input.Model = strings.TrimSpace(input.Model)
	if input.Model == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Model is required"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	snapshot, err := h.Service.UpdateChatBotModel(r.Context(), svc.UpdateChatBotModelCommand{UUID: uuid, UserID: userID, Model: input.Model})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Bot or enabled model").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) GetChatSnapshot(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	snapshot, err := h.Service.ChatSnapshotByUUID(r.Context(), uuidStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat snapshot"))
		return
	}
	json.NewEncoder(w).Encode(snapshotResponse(snapshot))
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

	chatSnapshots, err := h.Service.ChatSnapshotMetaByUserID(r.Context(), svc.SnapshotPageQuery{UserID: userID, Type: typ, Page: svc.PageWindow{Limit: pageSize, Offset: offset}})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve chat snapshots").WithDebugInfo(err.Error()))
		return
	}

	totalCount, err := h.Service.ChatSnapshotCountByUserIDAndType(r.Context(), userID, typ)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve snapshot count").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	data := make([]snapshotSummaryHTTPResponse, 0, len(chatSnapshots))
	for _, snapshot := range chatSnapshots {
		data = append(data, snapshotSummaryResponse(snapshot))
	}
	json.NewEncoder(w).Encode(snapshotPageHTTPResponse{Data: data, Page: page, PageSize: pageSize, Total: totalCount})
}

func (h *ChatSnapshotHandler) UpdateChatSnapshotMetaByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	}
	if err := DecodeJSON(r, &input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if err := h.Service.UpdateChatSnapshotMetadata(r.Context(), svc.UpdateSnapshotMetadataCommand{UUID: uuid, Title: input.Title, Summary: input.Summary, UserID: userID}); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update chat snapshot metadata").WithDebugInfo(err.Error()))
		return
	}

	snapshot, err := h.Service.ChatSnapshotByUUID(r.Context(), uuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	if err := h.Service.DeleteChatSnapshot(r.Context(), uuid, userID); err != nil {
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

	chatSnapshots, err := h.Service.ChatSnapshotSearch(r.Context(), userID, search)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to search chat snapshots").WithDebugInfo(err.Error()))
		return
	}

	response := make([]snapshotSearchHTTPResponse, 0, len(chatSnapshots))
	for _, snapshot := range chatSnapshots {
		response = append(response, snapshotSearchResponse(snapshot))
	}
	json.NewEncoder(w).Encode(response)
}
