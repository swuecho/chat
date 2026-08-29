package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
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

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (h *UserActiveChatSessionHandler) Register(router *mux.Router) {
	router.HandleFunc("/uuid/user_active_chat_session", h.GetUserActiveChatSessionHandler).Methods(http.MethodGet)
	router.HandleFunc("/uuid/user_active_chat_session", h.CreateOrUpdateUserActiveChatSessionHandler).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/active-sessions", h.GetAllWorkspaceActiveSessionsHandler).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{workspaceUuid}/active-session", h.GetWorkspaceActiveSessionHandler).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{workspaceUuid}/active-session", h.SetWorkspaceActiveSessionHandler).Methods(http.MethodPut)
}

func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	session, err := h.service.GetActiveSession(r.Context(), svc.GetActiveSessionQuery{UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID)))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get active chat session"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(activeSessionResponse(session)); err != nil {
		slog.Info("Failed to encode response", "error", err)
	}
}

func (h *UserActiveChatSessionHandler) CreateOrUpdateUserActiveChatSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	var reqBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	if !uuidRegex.MatchString(reqBody.ChatSessionUuid) {
		dto.RespondWithAPIError(w, dto.ErrChatSessionInvalid.WithMessage(
			fmt.Sprintf("invalid session UUID format: %s", reqBody.ChatSessionUuid)))
		return
	}

	slog.Info("Creating/updating active chat session", "userID", userID)

	session, err := h.service.UpsertActiveSession(r.Context(), svc.SetActiveSessionCommand{UserID: userID, SessionUUID: reqBody.ChatSessionUuid})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to create or update active chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(activeSessionResponse(session)); err != nil {
		slog.Info("Failed to encode response", "error", err)
	}
}

func (h *UserActiveChatSessionHandler) GetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	hasPermission, err := h.workspaceService.HasWorkspacePermission(ctx, workspaceUuid, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to check workspace permission"))
		return
	}
	if !hasPermission {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("access denied to workspace"))
		return
	}

	workspace, err := h.workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	session, err := h.service.GetActiveSession(ctx, svc.GetActiveSessionQuery{UserID: userID, WorkspaceID: &workspace.ID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Active Session").WithMessage("no active session for workspace"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaceActiveSessionHTTPResponse{SessionUUID: session.SessionUUID, WorkspaceUUID: workspaceUuid, UpdatedAt: session.UpdatedAt})
}

func (h *UserActiveChatSessionHandler) SetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	var requestBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	if !uuidRegex.MatchString(requestBody.ChatSessionUuid) {
		dto.RespondWithAPIError(w, dto.ErrChatSessionInvalid.WithMessage("invalid session UUID format"))
		return
	}

	activeSession, err := h.service.SetWorkspaceActiveSession(ctx, svc.SetWorkspaceActiveSessionCommand{
		UserID: userID, WorkspaceUUID: workspaceUuid, SessionUUID: requestBody.ChatSessionUuid,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaceActiveSessionHTTPResponse{SessionUUID: activeSession.SessionUUID, WorkspaceUUID: workspaceUuid, UpdatedAt: activeSession.UpdatedAt})
}

func (h *UserActiveChatSessionHandler) GetAllWorkspaceActiveSessionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	sessions, err := h.service.GetAllActiveSessions(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get workspace active sessions"))
		return
	}

	workspaces, err := h.workspaceService.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get workspaces"))
		return
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
