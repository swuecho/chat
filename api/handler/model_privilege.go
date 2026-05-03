package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type UserChatModelPrivilegeHandler struct {
	db *sqlc_queries.Queries
}

func NewUserChatModelPrivilegeHandler(db *sqlc_queries.Queries) *UserChatModelPrivilegeHandler {
	return &UserChatModelPrivilegeHandler{db: db}
}

func (h *UserChatModelPrivilegeHandler) Register(r *mux.Router) {
	r.HandleFunc("/admin/user_chat_model_privilege", h.ListUserChatModelPrivileges).Methods(http.MethodGet)
	r.HandleFunc("/admin/user_chat_model_privilege", h.CreateUserChatModelPrivilege).Methods(http.MethodPost)
	r.HandleFunc("/admin/user_chat_model_privilege/{id}", h.DeleteUserChatModelPrivilege).Methods(http.MethodDelete)
	r.HandleFunc("/admin/user_chat_model_privilege/{id}", h.UpdateUserChatModelPrivilege).Methods(http.MethodPut)
}

type ChatModelPrivilege struct {
	ID            int32  `json:"id"`
	FullName      string `json:"fullName"`
	UserEmail     string `json:"userEmail"`
	ChatModelName string `json:"chatModelName"`
	RateLimit     int32  `json:"rateLimit"`
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivileges(w http.ResponseWriter, r *http.Request) {
	userChatModelRows, err := h.db.ListUserChatModelPrivilegesRateLimit(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to list user chat model privileges"))
		return
	}

	output := lo.Map(userChatModelRows, func(r sqlc_queries.ListUserChatModelPrivilegesRateLimitRow, idx int) ChatModelPrivilege {
		return ChatModelPrivilege{
			ID:            r.ID,
			FullName:      r.FullName,
			UserEmail:     r.UserEmail,
			ChatModelName: r.ChatModelName,
			RateLimit:     r.RateLimit,
		}
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func (h *UserChatModelPrivilegeHandler) CreateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) {
	var input ChatModelPrivilege
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	if input.UserEmail == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("user email is required"))
		return
	}
	if input.ChatModelName == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("chat model name is required"))
		return
	}
	if input.RateLimit <= 0 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("rate limit must be positive").WithMessage(
			fmt.Sprintf("invalid rate limit: %d", input.RateLimit)))
		return
	}

	log.Printf("Creating chat model privilege for user %s with model %s", input.UserEmail, input.ChatModelName)

	user, err := h.db.GetAuthUserByEmail(r.Context(), input.UserEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrResourceNotFound("user").WithMessage(
				fmt.Sprintf("user with email %s not found", input.UserEmail)))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get user by email"))
		}
		return
	}

	chatModel, err := h.db.ChatModelByName(r.Context(), input.ChatModelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrChatModelNotFound.WithMessage(fmt.Sprintf("chat model %s not found", input.ChatModelName)))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get chat model"))
		}
		return
	}

	userChatModelPrivilege, err := h.db.CreateUserChatModelPrivilege(r.Context(), sqlc_queries.CreateUserChatModelPrivilegeParams{
		UserID:      user.ID,
		ChatModelID: chatModel.ID,
		RateLimit:   input.RateLimit,
		CreatedBy:   user.ID,
		UpdatedBy:   user.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model privilege"))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to create user chat model privilege"))
		}
		return
	}

	output := ChatModelPrivilege{
		ID:            userChatModelPrivilege.ID,
		UserEmail:     user.Email,
		ChatModelName: chatModel.Name,
		RateLimit:     userChatModelPrivilege.RateLimit,
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

	var input ChatModelPrivilege
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	if input.RateLimit <= 0 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("rate limit must be positive"))
		return
	}

	userChatModelPrivilege, err := h.db.UpdateUserChatModelPrivilege(r.Context(), sqlc_queries.UpdateUserChatModelPrivilegeParams{
		ID:        int32(id),
		RateLimit: input.RateLimit,
		UpdatedBy: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model privilege"))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to update user chat model privilege"))
		}
		return
	}
	output := ChatModelPrivilege{
		ID:            userChatModelPrivilege.ID,
		UserEmail:     input.UserEmail,
		ChatModelName: input.ChatModelName,
		RateLimit:     userChatModelPrivilege.RateLimit,
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

	if err := h.db.DeleteUserChatModelPrivilege(r.Context(), int32(id)); err != nil {
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

	privileges, err := h.db.ListUserChatModelPrivilegesByUserID(r.Context(), int32(userID))
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to list privileges for user"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(privileges)
}
