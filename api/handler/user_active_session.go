package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

func (h *UserActiveChatSessionHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/user_active_chat_session", endpoint(h.GetUserActiveChatSessionHandler)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/user_active_chat_session", endpoint(h.CreateOrUpdateUserActiveChatSessionHandler)).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/active-sessions", endpoint(h.GetAllWorkspaceActiveSessionsHandler)).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{workspaceUuid}/active-session", endpoint(h.GetWorkspaceActiveSessionHandler)).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{workspaceUuid}/active-session", endpoint(h.SetWorkspaceActiveSessionHandler)).Methods(http.MethodPut)
}

func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	session, err := h.service.GetActiveSession(r.Context(), svc.GetActiveSessionQuery{UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID))
		}
		return dto.WrapError(err, "failed to get active chat session")
	}
	return respondJSON(w, http.StatusOK, activeSessionResponse(session))
}

func (h *UserActiveChatSessionHandler) CreateOrUpdateUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var reqBody activeSessionRequest
	if err := DecodeJSON(r, &reqBody); err != nil {
		return dto.ErrValidationInvalidInput("failed to parse request body").WithDebugInfo(err.Error())
	}

	session, err := h.service.SetGlobalActiveSession(r.Context(), svc.SetGlobalActiveSessionCommand{UserID: userID, SessionUUID: reqBody.ChatSessionUUID})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusOK, activeSessionResponse(session))
}

func (h *UserActiveChatSessionHandler) GetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	session, err := h.service.GetWorkspaceActiveSession(ctx, svc.GetWorkspaceActiveSessionQuery{UserID: userID, WorkspaceUUID: workspaceUuid})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusOK, workspaceActiveSessionHTTPResponse{SessionUUID: session.SessionUUID, WorkspaceUUID: workspaceUuid, UpdatedAt: session.UpdatedAt})
}

func (h *UserActiveChatSessionHandler) SetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var requestBody activeSessionRequest
	if err := DecodeJSON(r, &requestBody); err != nil {
		return dto.ErrValidationInvalidInput("failed to parse request body").WithDebugInfo(err.Error())
	}

	activeSession, err := h.service.SetWorkspaceActiveSession(ctx, svc.SetWorkspaceActiveSessionCommand{
		UserID: userID, WorkspaceUUID: workspaceUuid, SessionUUID: requestBody.ChatSessionUUID,
	})
	if err != nil {
		return err
	}
	return respondJSON(w, http.StatusOK, workspaceActiveSessionHTTPResponse{SessionUUID: activeSession.SessionUUID, WorkspaceUUID: workspaceUuid, UpdatedAt: activeSession.UpdatedAt})
}

func (h *UserActiveChatSessionHandler) GetAllWorkspaceActiveSessionsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	sessions, err := h.service.GetAllActiveSessions(ctx, userID)
	if err != nil {
		return dto.WrapError(err, "failed to get workspace active sessions")
	}

	workspaces, err := h.workspaceService.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		return dto.WrapError(err, "failed to get workspaces")
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

	return respondJSON(w, http.StatusOK, response)
}
