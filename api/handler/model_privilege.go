package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type UserChatModelPrivilegeHandler struct {
	service *svc.ChatModelPrivilegeService
}

func NewUserChatModelPrivilegeHandler(service *svc.ChatModelPrivilegeService) *UserChatModelPrivilegeHandler {
	return &UserChatModelPrivilegeHandler{service: service}
}

func (h *UserChatModelPrivilegeHandler) Register(r *mux.Router) {
	r.HandleFunc("/admin/user_chat_model_privilege", h.ListUserChatModelPrivileges).Methods(http.MethodGet)
	r.HandleFunc("/admin/user_chat_model_privilege", h.CreateUserChatModelPrivilege).Methods(http.MethodPost)
	r.HandleFunc("/admin/user_chat_model_privilege/{id}", h.DeleteUserChatModelPrivilege).Methods(http.MethodDelete)
	r.HandleFunc("/admin/user_chat_model_privilege/{id}", h.UpdateUserChatModelPrivilege).Methods(http.MethodPut)
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivileges(w http.ResponseWriter, r *http.Request) {
	output, err := h.service.List(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to list user chat model privileges"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func (h *UserChatModelPrivilegeHandler) CreateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) {
	var request chatModelPrivilegeRequest
	if err := DecodeJSON(r, &request); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	if request.UserEmail == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("user email is required"))
		return
	}
	if request.ChatModelName == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("chat model name is required"))
		return
	}
	if request.RateLimit <= 0 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("rate limit must be positive").WithMessage(
			fmt.Sprintf("invalid rate limit: %d", request.RateLimit)))
		return
	}

	slog.Info("Creating chat model privilege", "userEmail", request.UserEmail, "chatModelName", request.ChatModelName)

	output, err := h.service.Create(r.Context(), svc.ChatModelPrivilege{
		UserEmail: request.UserEmail, ChatModelName: request.ChatModelName, RateLimit: request.RateLimit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model privilege"))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to create user chat model privilege"))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func (h *UserChatModelPrivilegeHandler) UpdateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	var request chatModelPrivilegeRequest
	if err := DecodeJSON(r, &request); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	if request.RateLimit <= 0 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("rate limit must be positive"))
		return
	}

	output, err := h.service.Update(r.Context(), int32(id), request.RateLimit, userID, request.UserEmail, request.ChatModelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model privilege"))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to update user chat model privilege"))
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func (h *UserChatModelPrivilegeHandler) DeleteUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	if err := h.service.Delete(r.Context(), int32(id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model privilege"))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to delete user chat model privilege"))
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivilegesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	privileges, err := h.service.ListByUserID(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to list privileges for user"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(privileges)
}
