package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatSessionHandler struct {
	service *svc.ChatSessionService
}

func NewChatSessionHandler(service *svc.ChatSessionService) *ChatSessionHandler {
	return &ChatSessionHandler{service: service}
}

func (h *ChatSessionHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	apicontract.RegisterJSON(router, registry, sessionOperation(http.MethodGet, "/chat_sessions/user",
		"listChatSessions", "List the authenticated user's chat sessions", http.StatusOK), h.getSimpleChatSessionsByUserID)
	apicontract.RegisterJSON(router, registry, sessionOperationWithUUID(http.MethodPut, "/uuid/chat_sessions/max_length/{uuid}",
		"updateSessionMaxLength", "Update a chat session token limit", http.StatusOK), h.updateSessionMaxLength)
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodPut, Path: "/uuid/chat_sessions/topic/{uuid}",
		OperationID: "updateChatSessionTopic", Summary: "Update a chat session topic",
		Tags: []string{"sessions"}, SuccessStatus: http.StatusOK,
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")},
		Security:   apicontract.BearerAuth(),
	}, h.updateChatSessionTopicByUUID)
	apicontract.RegisterJSON(router, registry, sessionOperationWithUUID(http.MethodGet, "/uuid/chat_sessions/{uuid}",
		"getChatSession", "Get a chat session", http.StatusOK), h.getChatSessionByUUID)
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodPut, Path: "/uuid/chat_sessions/{uuid}",
		OperationID: "createOrUpdateChatSession", Summary: "Create or update a chat session",
		Tags: []string{"sessions"}, SuccessStatus: http.StatusOK,
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")},
		Security:   apicontract.BearerAuth(),
	}, h.createOrUpdateChatSessionByUUID)
	apicontract.RegisterJSON(router, registry, sessionOperationWithUUID(http.MethodDelete, "/uuid/chat_sessions/{uuid}",
		"deleteChatSession", "Delete a chat session", http.StatusOK), h.deleteChatSessionByUUID)
	apicontract.RegisterJSON(router, registry, sessionOperation(http.MethodPost, "/uuid/chat_sessions",
		"createChatSession", "Create a chat session", http.StatusCreated), h.createChatSessionByUUID)
	apicontract.RegisterJSON(router, registry, sessionOperationWithUUID(http.MethodPost, "/uuid/chat_session_from_snapshot/{uuid}",
		"createChatSessionFromSnapshot", "Create a chat session from a snapshot", http.StatusCreated), h.createChatSessionFromSnapshot)
}

func sessionOperation(method, path, operationID, summary string, status int) apicontract.Operation {
	return apicontract.Operation{Method: method, Path: path, OperationID: operationID, Summary: summary,
		Tags: []string{"sessions"}, SuccessStatus: status, Security: apicontract.BearerAuth()}
}

func sessionOperationWithUUID(method, path, operationID, summary string, status int) apicontract.Operation {
	op := sessionOperation(method, path, operationID, summary, status)
	op.Parameters = []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")}
	return op
}

func (h *ChatSessionHandler) getChatSessionByUUID(r *http.Request, _ apicontract.NoBody) (dto.ChatSessionResponse, error) {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return dto.ChatSessionResponse{}, err
	}
	session, err := h.service.GetOwnedChatSession(r.Context(), svc.GetOwnedChatSessionQuery{UUID: uuid, UserID: userID})
	if err != nil {
		return dto.ChatSessionResponse{}, err
	}
	return dto.ChatSessionResponse{
		Uuid:            session.UUID,
		Topic:           session.Topic,
		MaxLength:       session.MaxLength,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
		ArtifactEnabled: session.ArtifactEnabled,
	}, nil
}

func (h *ChatSessionHandler) createChatSessionByUUID(r *http.Request, req createChatSessionRequest) (chatSessionHTTPResponse, error) {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}

	session, err := h.service.SaveSession(ctx, svc.SaveSessionCommand{
		SessionUUID: req.UUID, UserID: userID, Topic: req.Topic,
		MaxLength: dto.DefaultMaxLength, Temperature: dto.DefaultTemperature,
		Model: req.Model, MaxTokens: dto.DefaultMaxTokens, TopP: dto.DefaultTopP, N: dto.DefaultN,
		SummarizeMode: false, ExploreMode: false, ArtifactEnabled: false,
		EnsureSystemPrompt: true, DefaultSystemPrompt: req.DefaultSystemPrompt, ActivateGlobally: true,
	})
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}
	return chatSessionResponse(session), nil
}

func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(r *http.Request, sessionReq dto.UpdateChatSessionRequest) (chatSessionHTTPResponse, error) {
	pathUUID := mux.Vars(r)["uuid"]
	if sessionReq.Uuid != pathUUID {
		return chatSessionHTTPResponse{}, dto.ErrValidationInvalidInput("request uuid must match path uuid")
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = dto.DefaultMaxLength
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}

	command := svc.SaveSessionCommand{
		SessionUUID: sessionReq.Uuid, UserID: userID, Topic: sessionReq.Topic,
		MaxLength: sessionReq.MaxLength, Temperature: sessionReq.Temperature,
		Model: sessionReq.Model, TopP: sessionReq.TopP, N: sessionReq.N,
		MaxTokens:     sessionReq.MaxTokens,
		SummarizeMode: sessionReq.SummarizeMode, ArtifactEnabled: sessionReq.ArtifactEnabled,
		ExploreMode: sessionReq.ExploreMode, WorkspaceUUID: sessionReq.WorkspaceUUID,
	}

	session, err := h.service.SaveSession(ctx, command)
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}
	return chatSessionResponse(session), nil
}

func (h *ChatSessionHandler) deleteChatSessionByUUID(r *http.Request, _ apicontract.NoBody) (apicontract.NoBody, error) {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return apicontract.NoBody{}, err
	}

	if err := h.service.DeleteChatSessionByUUID(r.Context(), svc.DeleteChatSessionCommand{UUID: uuid, UserID: userID}); err != nil {
		return apicontract.NoBody{}, err
	}
	return apicontract.NoBody{}, nil
}

func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(r *http.Request, _ apicontract.NoBody) ([]dto.SimpleChatSession, error) {
	id, err := authenticatedUserID(r)
	if err != nil {
		return nil, err
	}
	sessions, err := h.service.GetSimpleChatSessionsByUserID(r.Context(), id)
	if err != nil {
		return nil, dto.ErrResourceNotFound("Chat sessions").WithDebugInfo(err.Error())
	}
	response := make([]dto.SimpleChatSession, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, dto.SimpleChatSession{
			Uuid: session.UUID, IsEdit: false, Title: session.Title,
			MaxLength: session.MaxLength, Temperature: session.Temperature,
			TopP: session.TopP, N: session.N, MaxTokens: session.MaxTokens,
			Model:         session.Model,
			SummarizeMode: session.SummarizeMode, ArtifactEnabled: session.ArtifactEnabled,
			WorkspaceUuid: session.WorkspaceUUID,
		})
	}
	return response, nil
}

func (h *ChatSessionHandler) updateChatSessionTopicByUUID(r *http.Request, req updateSessionTopicRequest) (chatSessionHTTPResponse, error) {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}
	session, err := h.service.UpdateChatSessionTopicByUUID(r.Context(), uuid, userID, req.Topic)
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}
	return chatSessionResponse(session), nil
}

func (h *ChatSessionHandler) updateSessionMaxLength(r *http.Request, req updateSessionMaxLengthRequest) (chatSessionHTTPResponse, error) {
	uuid := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}
	updatedSession, err := h.service.UpdateSessionMaxLength(r.Context(), svc.UpdateSessionMaxLengthCommand{UUID: uuid, UserID: userID, MaxLength: req.MaxLength})
	if err != nil {
		return chatSessionHTTPResponse{}, err
	}
	return chatSessionResponse(updatedSession), nil
}

func (h *ChatSessionHandler) createChatSessionFromSnapshot(r *http.Request, _ apicontract.NoBody) (sessionCreatedHTTPResponse, error) {
	snapshotUUID := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return sessionCreatedHTTPResponse{}, err
	}

	result, err := h.service.CreateSessionFromSnapshot(r.Context(), svc.CreateSessionFromSnapshotCommand{
		SnapshotUUID: snapshotUUID,
		UserID:       userID,
	})
	if err != nil {
		return sessionCreatedHTTPResponse{}, err
	}
	return sessionCreatedHTTPResponse{SessionUUID: result.SessionUUID}, nil
}
