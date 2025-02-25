package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rotisserie/eris"
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
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error listing user chat model privileges").Error(), err)
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid user chat model privilege ID"))
		return
	}

	userChatModelPrivilege, err := h.db.UserChatModelPrivilegeByID(r.Context(), int32(id))
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting user chat model privilege").Error(), err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userChatModelPrivilege)
}

func (h *UserChatModelPrivilegeHandler) CreateUserChatModelPrivilege(w http.ResponseWriter, r *http.Request) {
	var input ChatModelPrivilege
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Failed to parse request body").Error(), err)
		return
	}

	user, err := h.db.GetAuthUserByEmail(r.Context(), input.UserEmail)

	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Failed to get user by email").Error(), err)
	}
	chatModel, err := h.db.ChatModelByName(r.Context(), input.ChatModelName)
	if err != nil {
		RespondWithServerErrorRepsonse(w, ErrModelNotFound.WithDetails(chatModel.Name))

	}
	log.Printf("%+v\n", chatModel)

	userChatModelPrivilege, err := h.db.CreateUserChatModelPrivilege(r.Context(), sqlc_queries.CreateUserChatModelPrivilegeParams{
		UserID:      user.ID,
		ChatModelID: chatModel.ID,
		RateLimit:   input.RateLimit,
		CreatedBy:   user.ID,
		UpdatedBy:   user.ID,
	})

	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error creating user chat model privilege").Error(), err)
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid user chat model privilege ID"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithErrorMessage(w, http.StatusUnauthorized, "Unauthorized", err)
	}

	var input ChatModelPrivilege
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Failed to parse request body").Error(), err)
		return
	}

	userChatModelPrivilege, err := h.db.UpdateUserChatModelPrivilege(r.Context(), sqlc_queries.UpdateUserChatModelPrivilegeParams{
		ID:        int32(id),
		RateLimit: input.RateLimit,
		UpdatedBy: userID,
	})

	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error updating user chat model privilege").Error(), err)
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid user chat model privilege ID"))
		return
	}

	err = h.db.DeleteUserChatModelPrivilege(r.Context(), int32(id))
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error deleting user chat model privilege").Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserChatModelPrivilegeHandler) UserChatModelPrivilegeByUserAndModelID(w http.ResponseWriter, r *http.Request) {
	_, err := getUserID(r.Context())
	if err != nil {
		RespondWithErrorMessage(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	var input struct {
		UserID      int32
		ChatModelID int32
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Failed to parse request body").Error(), err)
		return
	}

	userChatModelPrivilege, err := h.db.UserChatModelPrivilegeByUserAndModelID(r.Context(),
		sqlc_queries.UserChatModelPrivilegeByUserAndModelIDParams{
			UserID:      input.UserID,
			ChatModelID: input.ChatModelID,
		})

	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting user chat model privilege").Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userChatModelPrivilege)
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivilegesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithErrorMessage(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	privileges, err := h.db.ListUserChatModelPrivilegesByUserID(r.Context(), int32(userID))

	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error listing privileges for user").Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(privileges)
}
