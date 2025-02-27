package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type UserChatModelPrivilegeHandler struct {
	db *sqlc_queries.Queries
}

func NewUserChatModelPrivilegeHandler(db *sqlc_queries.Queries) *UserChatModelPrivilegeHandler {
	return &UserChatModelPrivilegeHandler{
		db: db,
	}
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
	// TODO: check user is super_user
	userChatModelRows, err := h.db.ListUserChatModelPrivilegesRateLimit(r.Context())
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to list user chat model privileges"))
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

func (h *UserChatModelPrivilegeHandler) UserChatModelPrivilegeByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	userChatModelPrivilege, err := h.db.UserChatModelPrivilegeByID(r.Context(), int32(id))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get user chat model privilege"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userChatModelPrivilege)
}

func (h *UserChatModelPrivilegeHandler) CreateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) {
	var input ChatModelPrivilege
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	user, err := h.db.GetAuthUserByEmail(r.Context(), input.UserEmail)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("user with email "+input.UserEmail))
		return
	}

	chatModel, err := h.db.ChatModelByName(r.Context(), input.ChatModelName)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model "+input.ChatModelName))
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
		RespondWithAPIError(w, WrapError(err, "failed to create user chat model privilege"))
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
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	var input ChatModelPrivilege
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	userChatModelPrivilege, err := h.db.UpdateUserChatModelPrivilege(r.Context(), sqlc_queries.UpdateUserChatModelPrivilegeParams{
		ID:        int32(id),
		RateLimit: input.RateLimit,
		UpdatedBy: userID,
	})

	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to update user chat model privilege"))
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
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	err = h.db.DeleteUserChatModelPrivilege(r.Context(), int32(id))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to delete user chat model privilege"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserChatModelPrivilegeHandler) UserChatModelPrivilegeByUserAndModelID(w http.ResponseWriter, r *http.Request) {
	_, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	var input struct {
		UserID      int32
		ChatModelID int32
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	userChatModelPrivilege, err := h.db.UserChatModelPrivilegeByUserAndModelID(r.Context(),
		sqlc_queries.UserChatModelPrivilegeByUserAndModelIDParams{
			UserID:      input.UserID,
			ChatModelID: input.ChatModelID,
		})

	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get user chat model privilege"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userChatModelPrivilege)
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivilegesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	privileges, err := h.db.ListUserChatModelPrivilegesByUserID(r.Context(), int32(userID))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to list privileges for user"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(privileges)
}
