package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type UserActiveChatSessionHandler struct {
	service          *svc.UserActiveChatSessionService
	workspaceService *svc.ChatWorkspaceService
}

type activeSessionHTTPResponse struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"userId"`
	SessionUUID string    `json:"chatSessionUuid"`
	WorkspaceID *int32    `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type workspaceActiveSessionHTTPResponse struct {
	WorkspaceUUID string    `json:"workspaceUuid"`
	SessionUUID   string    `json:"chatSessionUuid"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func activeSessionResponse(session svc.ActiveSession) activeSessionHTTPResponse {
	return activeSessionHTTPResponse{ID: session.ID, UserID: session.UserID, SessionUUID: session.SessionUUID,
		WorkspaceID: session.WorkspaceID, CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt}
}

func NewUserActiveChatSessionHandler(service *svc.UserActiveChatSessionService, workspaceService *svc.ChatWorkspaceService) *UserActiveChatSessionHandler {
	return &UserActiveChatSessionHandler{
		service: service, workspaceService: workspaceService,
	}
}

func (h *UserActiveChatSessionHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	apicontract.RegisterJSON(router, registry, activeSessionOperation(http.MethodGet, "/uuid/user_active_chat_session", "getActiveChatSession", "Get the global active chat session", http.StatusOK), h.GetUserActiveChatSessionHandler)
	apicontract.RegisterJSON(router, registry, activeSessionOperation(http.MethodPut, "/uuid/user_active_chat_session", "setActiveChatSession", "Set the global active chat session", http.StatusOK), h.CreateOrUpdateUserActiveChatSessionHandler)
	apicontract.RegisterJSON(router, registry, activeSessionOperation(http.MethodGet, "/workspaces/active-sessions", "listWorkspaceActiveSessions", "List workspace active sessions", http.StatusOK), h.GetAllWorkspaceActiveSessionsHandler)
	apicontract.RegisterJSON(router, registry, workspaceActiveSessionOperation(http.MethodGet, "getWorkspaceActiveSession", "Get a workspace active session"), h.GetWorkspaceActiveSessionHandler)
	apicontract.RegisterJSON(router, registry, workspaceActiveSessionOperation(http.MethodPut, "setWorkspaceActiveSession", "Set a workspace active session"), h.SetWorkspaceActiveSessionHandler)
}

func activeSessionOperation(method, path, operationID, summary string, status int) apicontract.Operation {
	return apicontract.Operation{Method: method, Path: path, OperationID: operationID, Summary: summary,
		Tags: []string{"active-sessions"}, SuccessStatus: status, Security: apicontract.BearerAuth()}
}

func workspaceActiveSessionOperation(method, operationID, summary string) apicontract.Operation {
	op := activeSessionOperation(method, "/workspaces/{workspaceUuid}/active-session", operationID, summary, http.StatusOK)
	op.Parameters = []apicontract.Parameter{apicontract.UUIDPathParameter("workspaceUuid")}
	return op
}

func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(r *http.Request, _ apicontract.NoBody) (activeSessionHTTPResponse, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return activeSessionHTTPResponse{}, err
	}

	session, err := h.service.GetActiveSession(r.Context(), svc.GetActiveSessionQuery{UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return activeSessionHTTPResponse{}, dto.ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID))
		}
		return activeSessionHTTPResponse{}, dto.WrapError(err, "failed to get active chat session")
	}
	return activeSessionResponse(session), nil
}

func (h *UserActiveChatSessionHandler) CreateOrUpdateUserActiveChatSessionHandler(r *http.Request, reqBody activeSessionRequest) (activeSessionHTTPResponse, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return activeSessionHTTPResponse{}, err
	}

	session, err := h.service.SetGlobalActiveSession(r.Context(), svc.SetGlobalActiveSessionCommand{UserID: userID, SessionUUID: reqBody.ChatSessionUUID})
	if err != nil {
		return activeSessionHTTPResponse{}, err
	}
	return activeSessionResponse(session), nil
}

func (h *UserActiveChatSessionHandler) GetWorkspaceActiveSessionHandler(r *http.Request, _ apicontract.NoBody) (workspaceActiveSessionHTTPResponse, error) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := authenticatedUserID(r)
	if err != nil {
		return workspaceActiveSessionHTTPResponse{}, err
	}

	session, err := h.service.GetWorkspaceActiveSession(ctx, svc.GetWorkspaceActiveSessionQuery{UserID: userID, WorkspaceUUID: workspaceUuid})
	if err != nil {
		return workspaceActiveSessionHTTPResponse{}, err
	}
	return workspaceActiveSessionHTTPResponse{SessionUUID: session.SessionUUID, WorkspaceUUID: workspaceUuid, UpdatedAt: session.UpdatedAt}, nil
}

func (h *UserActiveChatSessionHandler) SetWorkspaceActiveSessionHandler(r *http.Request, requestBody activeSessionRequest) (workspaceActiveSessionHTTPResponse, error) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := authenticatedUserID(r)
	if err != nil {
		return workspaceActiveSessionHTTPResponse{}, err
	}

	activeSession, err := h.service.SetWorkspaceActiveSession(ctx, svc.SetWorkspaceActiveSessionCommand{
		UserID: userID, WorkspaceUUID: workspaceUuid, SessionUUID: requestBody.ChatSessionUUID,
	})
	if err != nil {
		return workspaceActiveSessionHTTPResponse{}, err
	}
	return workspaceActiveSessionHTTPResponse{SessionUUID: activeSession.SessionUUID, WorkspaceUUID: workspaceUuid, UpdatedAt: activeSession.UpdatedAt}, nil
}

func (h *UserActiveChatSessionHandler) GetAllWorkspaceActiveSessionsHandler(r *http.Request, _ apicontract.NoBody) ([]workspaceActiveSessionHTTPResponse, error) {
	ctx := r.Context()

	userID, err := authenticatedUserID(r)
	if err != nil {
		return nil, err
	}

	sessions, err := h.service.GetAllActiveSessions(ctx, userID)
	if err != nil {
		return nil, dto.WrapError(err, "failed to get workspace active sessions")
	}

	workspaces, err := h.workspaceService.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		return nil, dto.WrapError(err, "failed to get workspaces")
	}

	workspaceMap := make(map[int32]string)
	for _, workspace := range workspaces {
		workspaceMap[workspace.ID] = workspace.UUID
	}

	response := make([]workspaceActiveSessionHTTPResponse, 0, len(sessions))
	for _, session := range sessions {
		if session.WorkspaceID != nil {
			if workspaceUuid, exists := workspaceMap[*session.WorkspaceID]; exists {
				response = append(response, workspaceActiveSessionHTTPResponse{WorkspaceUUID: workspaceUuid, SessionUUID: session.SessionUUID, UpdatedAt: session.UpdatedAt})
			}
		}
	}

	return response, nil
}
