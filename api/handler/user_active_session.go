package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type UserActiveChatSessionHandler struct {
	service *svc.UserActiveChatSessionService
}

func NewUserActiveChatSessionHandler(sqlc_q *sqlc_queries.Queries) *UserActiveChatSessionHandler {
	return &UserActiveChatSessionHandler{
		service: svc.NewUserActiveChatSessionService(sqlc_q),
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

	session, err := h.service.GetActiveSession(r.Context(), userID, nil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID)))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get active chat session"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		log.Printf("Failed to encode response: %v", err)
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

	log.Printf("Creating/updating active chat session for user %d", userID)

	session, err := h.service.UpsertActiveSession(r.Context(), userID, nil, reqBody.ChatSessionUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to create or update active chat session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		log.Printf("Failed to encode response: %v", err)
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

	workspaceService := svc.NewChatWorkspaceService(h.service.Q())
	hasPermission, err := workspaceService.HasWorkspacePermission(ctx, workspaceUuid, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to check workspace permission"))
		return
	}
	if !hasPermission {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("access denied to workspace"))
		return
	}

	workspace, err := workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	session, err := h.service.GetActiveSession(ctx, userID, &workspace.ID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Active Session").WithMessage("no active session for workspace"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

func (h *UserActiveChatSessionHandler) SetWorkspaceActiveSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceUuid := mux.Vars(r)["workspaceUuid"]

	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	workspaceService := svc.NewChatWorkspaceService(h.service.Q())
	hasPermission, err := workspaceService.HasWorkspacePermission(ctx, workspaceUuid, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to check workspace permission"))
		return
	}
	if !hasPermission {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("access denied to workspace"))
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

	workspace, err := workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	sessionService := svc.NewChatSessionService(h.service.Q())
	session, err := sessionService.GetChatSessionByUUID(ctx, requestBody.ChatSessionUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat Session").WithMessage("chat session not found"))
		return
	}
	if !session.WorkspaceID.Valid || session.WorkspaceID.Int32 != workspace.ID {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("session does not belong to workspace"))
		return
	}

	activeSession, err := h.service.UpsertActiveSession(ctx, userID, &workspace.ID, requestBody.ChatSessionUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to set workspace active session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatSessionUuid": activeSession.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       activeSession.UpdatedAt,
	})
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

	workspaceService := svc.NewChatWorkspaceService(h.service.Q())
	workspaces, err := workspaceService.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get workspaces"))
		return
	}

	workspaceMap := make(map[int32]string)
	for _, workspace := range workspaces {
		workspaceMap[workspace.ID] = workspace.Uuid
	}

	var response []map[string]interface{}
	for _, session := range sessions {
		if session.WorkspaceID.Valid {
			if workspaceUuid, exists := workspaceMap[session.WorkspaceID.Int32]; exists {
				response = append(response, map[string]interface{}{
					"workspaceUuid":   workspaceUuid,
					"chatSessionUuid": session.ChatSessionUuid,
					"updatedAt":       session.UpdatedAt,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
