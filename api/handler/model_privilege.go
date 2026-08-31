package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type UserChatModelPrivilegeHandler struct {
	service *svc.ChatModelPrivilegeService
}

func NewUserChatModelPrivilegeHandler(service *svc.ChatModelPrivilegeService) *UserChatModelPrivilegeHandler {
	return &UserChatModelPrivilegeHandler{service: service}
}

func (h *UserChatModelPrivilegeHandler) Register(r *mux.Router, registry *apicontract.Registry) {
	r.HandleFunc("/admin/user_chat_model_privilege", endpoint(h.ListUserChatModelPrivileges)).Methods(http.MethodGet)
	r.HandleFunc("/admin/user_chat_model_privilege", endpoint(h.CreateUserChatModelPrivilege)).Methods(http.MethodPost)
	r.HandleFunc("/admin/user_chat_model_privilege/{id}", endpoint(h.DeleteUserChatModelPrivilege)).Methods(http.MethodDelete)
	r.HandleFunc("/admin/user_chat_model_privilege/{id}", endpoint(h.UpdateUserChatModelPrivilege)).Methods(http.MethodPut)
	security := apicontract.BearerAuth()
	base := apicontract.Operation{Path: "/admin/user_chat_model_privilege", Tags: []string{"Admin model privileges"}, Security: security}
	base.Method, base.OperationID, base.Summary, base.SuccessStatus = http.MethodGet, "listUserChatModelPrivileges", "List user model privileges", http.StatusOK
	apicontract.DocumentJSON[apicontract.NoBody, []svc.ChatModelPrivilege](registry, base)
	base.Method, base.OperationID, base.Summary = http.MethodPost, "createUserChatModelPrivilege", "Create a user model privilege"
	apicontract.DocumentJSON[chatModelPrivilegeRequest, svc.ChatModelPrivilege](registry, base)
	withID := []apicontract.Parameter{apicontract.PositiveIntegerPathParameter("id")}
	apicontract.DocumentJSON[chatModelPrivilegeRequest, svc.ChatModelPrivilege](registry, apicontract.Operation{Method: http.MethodPut, Path: "/admin/user_chat_model_privilege/{id}", OperationID: "updateUserChatModelPrivilege", Summary: "Update a user model privilege", Tags: base.Tags, SuccessStatus: http.StatusOK, Security: security, Parameters: withID})
	apicontract.DocumentJSON[apicontract.NoBody, apicontract.NoBody](registry, apicontract.Operation{Method: http.MethodDelete, Path: "/admin/user_chat_model_privilege/{id}", OperationID: "deleteUserChatModelPrivilege", Summary: "Delete a user model privilege", Tags: base.Tags, SuccessStatus: http.StatusNoContent, Security: security, Parameters: withID})
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivileges(w http.ResponseWriter, r *http.Request) error {
	output, err := h.service.List(r.Context())
	if err != nil {
		return dto.WrapError(err, "failed to list user chat model privileges")
	}
	return respondJSON(w, http.StatusOK, output)
}

func (h *UserChatModelPrivilegeHandler) CreateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) error {
	var request chatModelPrivilegeRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("failed to parse request body").WithDebugInfo(err.Error())
	}

	if request.UserEmail == "" {
		return dto.ErrValidationInvalidInput("user email is required")
	}

	output, err := h.service.Create(r.Context(), svc.ChatModelPrivilege{
		UserEmail: request.UserEmail, ChatModelName: request.ChatModelName, RateLimit: request.RateLimit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.ErrResourceNotFound("chat model privilege")
		}
		return dto.WrapError(err, "failed to create user chat model privilege")
	}
	return respondJSON(w, http.StatusOK, output)
}

func (h *UserChatModelPrivilegeHandler) UpdateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var request chatModelPrivilegeRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("failed to parse request body").WithDebugInfo(err.Error())
	}

	output, err := h.service.Update(r.Context(), id, request.RateLimit, userID, request.UserEmail, request.ChatModelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.ErrResourceNotFound("chat model privilege")
		}
		return dto.WrapError(err, "failed to update user chat model privilege")
	}
	return respondJSON(w, http.StatusOK, output)
}

func (h *UserChatModelPrivilegeHandler) DeleteUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.ErrResourceNotFound("chat model privilege")
		}
		return dto.WrapError(err, "failed to delete user chat model privilege")
	}
	return noContent(w)
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivilegesByUserID(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	privileges, err := h.service.ListByUserID(r.Context(), userID)
	if err != nil {
		return dto.WrapError(err, "failed to list privileges for user")
	}
	return respondJSON(w, http.StatusOK, privileges)
}
