package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
)

type ChatSnapshotHandler struct {
	Service *svc.ChatSnapshotService
}

func NewChatSnapshotHandler(service *svc.ChatSnapshotService) *ChatSnapshotHandler {
	return &ChatSnapshotHandler{Service: service}
}

func (h *ChatSnapshotHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	router.HandleFunc("/uuid/chat_snapshot/all", endpoint(h.ChatSnapshotMetaByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", endpoint(h.GetChatSnapshot)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", endpoint(h.CreateChatSnapshot)).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", endpoint(h.UpdateChatSnapshotMetaByUUID)).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_snapshot/{uuid}", endpoint(h.DeleteChatSnapshot)).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_snapshot_search", endpoint(h.ChatSnapshotSearch)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_bot/{uuid}", endpoint(h.CreateChatBot)).Methods(http.MethodPost)
	router.HandleFunc("/uuid/chat_bot/{uuid}/model", endpoint(h.UpdateChatBotModel)).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_bot/{uuid}/settings", endpoint(h.UpdateChatBotSettings)).Methods(http.MethodPut)

	secured := func(op apicontract.Operation) apicontract.Operation {
		op.Tags = []string{"Snapshots"}
		op.Security = apicontract.BearerAuth()
		return op
	}
	uuidParameter := []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")}
	maxPageSize := validation.MaxPageSize
	apicontract.DocumentJSON[apicontract.NoBody, snapshotListHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/uuid/chat_snapshot/all", OperationID: "listChatSnapshots", Summary: "List chat snapshots", SuccessStatus: http.StatusOK, Parameters: []apicontract.Parameter{
		apicontract.StringQueryParameter("type", "snapshot", "chatbot"),
		apicontract.IntegerQueryParameter("limit", 1, &maxPageSize), apicontract.IntegerQueryParameter("offset", 0, nil),
	}}))
	apicontract.DocumentJSON[apicontract.NoBody, snapshotHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/uuid/chat_snapshot/{uuid}", OperationID: "getChatSnapshot", Summary: "Get a chat snapshot", SuccessStatus: http.StatusOK, Parameters: uuidParameter}))
	apicontract.DocumentJSON[apicontract.NoBody, uuidHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodPost, Path: "/uuid/chat_snapshot/{uuid}", OperationID: "createChatSnapshot", Summary: "Create a chat snapshot", SuccessStatus: http.StatusCreated, Parameters: uuidParameter}))
	apicontract.DocumentJSON[updateSnapshotMetadataRequest, snapshotHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodPut, Path: "/uuid/chat_snapshot/{uuid}", OperationID: "updateChatSnapshot", Summary: "Update chat snapshot metadata", SuccessStatus: http.StatusOK, Parameters: uuidParameter}))
	apicontract.DocumentJSON[apicontract.NoBody, apicontract.NoBody](registry, secured(apicontract.Operation{Method: http.MethodDelete, Path: "/uuid/chat_snapshot/{uuid}", OperationID: "deleteChatSnapshot", Summary: "Delete a chat snapshot", SuccessStatus: http.StatusOK, Parameters: uuidParameter}))
	apicontract.DocumentJSON[apicontract.NoBody, []snapshotSearchHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/uuid/chat_snapshot_search", OperationID: "searchChatSnapshots", Summary: "Search chat snapshots", SuccessStatus: http.StatusOK, Parameters: []apicontract.Parameter{apicontract.StringQueryParameter("search")}}))
	apicontract.DocumentJSON[apicontract.NoBody, uuidHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodPost, Path: "/uuid/chat_bot/{uuid}", OperationID: "createChatBot", Summary: "Create a chat bot", SuccessStatus: http.StatusCreated, Parameters: uuidParameter}))
	apicontract.DocumentJSON[updateBotModelRequest, snapshotHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodPut, Path: "/uuid/chat_bot/{uuid}/model", OperationID: "updateChatBotModel", Summary: "Update a chat bot model", SuccessStatus: http.StatusOK, Parameters: uuidParameter}))
	apicontract.DocumentJSON[updateBotSettingsRequest, snapshotHTTPResponse](registry, secured(apicontract.Operation{Method: http.MethodPut, Path: "/uuid/chat_bot/{uuid}/settings", OperationID: "updateChatBotSettings", Summary: "Update chat bot settings", SuccessStatus: http.StatusOK, Parameters: uuidParameter}))
}

func (h *ChatSnapshotHandler) CreateChatSnapshot(w http.ResponseWriter, r *http.Request) error {
	chatSessionUuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	uuid, err := h.Service.CreateChatSnapshot(r.Context(), chatSessionUuid, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat snapshot")
	}
	return respondJSON(w, http.StatusCreated, uuidHTTPResponse{UUID: uuid})
}

func (h *ChatSnapshotHandler) CreateChatBot(w http.ResponseWriter, r *http.Request) error {
	chatSessionUuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	uuid, err := h.Service.CreateChatBot(r.Context(), chatSessionUuid, userID)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to create chat bot").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusCreated, uuidHTTPResponse{UUID: uuid})
}

func (h *ChatSnapshotHandler) UpdateChatBotSettings(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	var input updateBotSettingsRequest
	if err := DecodeJSON(r, &input); err != nil {
		return dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error())
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	snapshot, err := h.Service.UpdateChatBotSettings(r.Context(), svc.UpdateChatBotSettingsCommand{UUID: uuid, UserID: userID, Title: input.Title, Summary: input.Summary, Model: input.Model})
	if err != nil {
		return dto.ErrResourceNotFound("Bot or enabled model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) UpdateChatBotModel(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	var input updateBotModelRequest
	if err := DecodeJSON(r, &input); err != nil {
		return dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error())
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	snapshot, err := h.Service.UpdateChatBotModel(r.Context(), svc.UpdateChatBotModelCommand{UUID: uuid, UserID: userID, Model: input.Model})
	if err != nil {
		return dto.ErrResourceNotFound("Bot or enabled model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) GetChatSnapshot(w http.ResponseWriter, r *http.Request) error {
	uuidStr := mux.Vars(r)["uuid"]
	snapshot, err := h.Service.ChatSnapshotByUUID(r.Context(), uuidStr)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat snapshot")
	}
	return respondJSON(w, http.StatusOK, snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) ChatSnapshotMetaByUserID(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	typ := r.URL.Query().Get("type")
	page, err := httpx.ParsePage(r)
	if err != nil {
		return err
	}

	chatSnapshots, err := h.Service.ChatSnapshotMetaByUserID(r.Context(), svc.SnapshotPageQuery{UserID: userID, Type: typ, Page: svc.PageWindow{Limit: page.Limit, Offset: page.Offset}})
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to retrieve chat snapshots").WithDebugInfo(err.Error())
	}

	totalCount, err := h.Service.ChatSnapshotCountByUserIDAndType(r.Context(), userID, typ)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to retrieve snapshot count").WithDebugInfo(err.Error())
	}
	data := make([]snapshotSummaryHTTPResponse, 0, len(chatSnapshots))
	for _, snapshot := range chatSnapshots {
		data = append(data, snapshotSummaryResponse(snapshot))
	}
	return respondJSON(w, http.StatusOK, pageResponse(data, totalCount, page))
}

func (h *ChatSnapshotHandler) UpdateChatSnapshotMetaByUUID(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	var input updateSnapshotMetadataRequest
	if err := DecodeJSON(r, &input); err != nil {
		return dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error())
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	if err := h.Service.UpdateChatSnapshotMetadata(r.Context(), svc.UpdateSnapshotMetadataCommand{UUID: uuid, Title: input.Title, Summary: input.Summary, UserID: userID}); err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to update chat snapshot metadata").WithDebugInfo(err.Error())
	}

	snapshot, err := h.Service.ChatSnapshotByUUID(r.Context(), uuid)
	if err != nil {
		return dto.ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, snapshotResponse(snapshot))
}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(w http.ResponseWriter, r *http.Request) error {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	if err := h.Service.DeleteChatSnapshot(r.Context(), uuid, userID); err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to delete chat snapshot").WithDebugInfo(err.Error())
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatSnapshotHandler) ChatSnapshotSearch(w http.ResponseWriter, r *http.Request) error {
	search := r.URL.Query().Get("search")
	if search == "" {
		return respondJSON(w, http.StatusOK, []snapshotSearchHTTPResponse{})
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	chatSnapshots, err := h.Service.ChatSnapshotSearch(r.Context(), userID, search)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to search chat snapshots").WithDebugInfo(err.Error())
	}

	response := make([]snapshotSearchHTTPResponse, 0, len(chatSnapshots))
	for _, snapshot := range chatSnapshots {
		response = append(response, snapshotSearchResponse(snapshot))
	}
	return respondJSON(w, http.StatusOK, response)
}
