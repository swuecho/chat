package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// UserChatModelPrivilegeHandler handles requests related to user chat model privileges
type UserChatModelPrivilegeHandler struct {
	db *sqlc_queries.Queries
}

// NewUserChatModelPrivilegeHandler creates a new handler instance
func NewUserChatModelPrivilegeHandler(db *sqlc_queries.Queries) *UserChatModelPrivilegeHandler {
	return &UserChatModelPrivilegeHandler{
		db: db,
	}
}

// Register sets up the handler routes
func (h *UserChatModelPrivilegeHandler) Register(r *gin.RouterGroup) {
	r.GET("/admin/user_chat_model_privilege", h.ListUserChatModelPrivileges)
	r.POST("/admin/user_chat_model_privilege", h.CreateUserChatModelPrivilege)
	r.DELETE("/admin/user_chat_model_privilege/:id", h.DeleteUserChatModelPrivilege)
	r.PUT("/admin/user_chat_model_privilege/:id", h.UpdateUserChatModelPrivilege)
}

type ChatModelPrivilege struct {
	ID            int32  `json:"id"`
	FullName      string `json:"fullName"`
	UserEmail     string `json:"userEmail"`
	ChatModelName string `json:"chatModelName"`
	RateLimit     int32  `json:"rateLimit"`
}

// ListUserChatModelPrivileges handles GET requests to list all user chat model privileges
func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivileges(c *gin.Context) {
	// TODO: check user is super_user
	userChatModelRows, err := h.db.ListUserChatModelPrivilegesRateLimit(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to list user chat model privileges"))
		return
	}

	log.Printf("Listing user chat model privileges")
	output := lo.Map(userChatModelRows, func(r sqlc_queries.ListUserChatModelPrivilegesRateLimitRow, idx int) ChatModelPrivilege {
		return ChatModelPrivilege{
			ID:            r.ID,
			FullName:      r.FullName,
			UserEmail:     r.UserEmail,
			ChatModelName: r.ChatModelName,
			RateLimit:     r.RateLimit,
		}
	})

	c.JSON(http.StatusOK, output)
}

func (h *UserChatModelPrivilegeHandler) UserChatModelPrivilegeByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	userChatModelPrivilege, err := h.db.UserChatModelPrivilegeByID(c.Request.Context(), int32(id))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to get user chat model privilege"))
		return
	}
	c.JSON(http.StatusOK, userChatModelPrivilege)
}

// CreateUserChatModelPrivilege handles POST requests to create a new user chat model privilege
func (h *UserChatModelPrivilegeHandler) CreateUserChatModelPrivilege(c *gin.Context) {
	var input ChatModelPrivilege
	err := c.ShouldBindJSON(&input)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	// Validate input
	if input.UserEmail == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("user email is required"))
		return
	}
	if input.ChatModelName == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("chat model name is required"))
		return
	}
	if input.RateLimit <= 0 {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("rate limit must be positive").WithMessage(
			fmt.Sprintf("invalid rate limit: %d", input.RateLimit)))
		return
	}

	log.Printf("Creating chat model privilege for user %s with model %s",
		input.UserEmail, input.ChatModelName)

	user, err := h.db.GetAuthUserByEmail(c.Request.Context(), input.UserEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrResourceNotFound("user").WithMessage(
				fmt.Sprintf("user with email %s not found", input.UserEmail)))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to get user by email"))
		}
		return
	}

	chatModel, err := h.db.ChatModelByName(c.Request.Context(), input.ChatModelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrChatModelNotFound.WithMessage(fmt.Sprintf("chat model %s not found", input.ChatModelName)))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to get chat model"))
		}
		return
	}

	userChatModelPrivilege, err := h.db.CreateUserChatModelPrivilege(c.Request.Context(), sqlc_queries.CreateUserChatModelPrivilegeParams{
		UserID:      user.ID,
		ChatModelID: chatModel.ID,
		RateLimit:   input.RateLimit,
		CreatedBy:   user.ID,
		UpdatedBy:   user.ID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrResourceNotFound("chat model privilege"))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to create user chat model privilege"))
		}
		return
	}

	output := ChatModelPrivilege{
		ID:            userChatModelPrivilege.ID,
		UserEmail:     user.Email,
		ChatModelName: chatModel.Name,
		RateLimit:     userChatModelPrivilege.RateLimit,
	}
	c.JSON(http.StatusOK, output)
}

// UpdateUserChatModelPrivilege handles PUT requests to update a user chat model privilege
func (h *UserChatModelPrivilegeHandler) UpdateUserChatModelPrivilege(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	var input ChatModelPrivilege
	err = c.ShouldBindJSON(&input)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	// Validate input
	if input.RateLimit <= 0 {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("rate limit must be positive"))
		return
	}

	log.Printf("Updating chat model privilege %d for user %d", id, userID)

	userChatModelPrivilege, err := h.db.UpdateUserChatModelPrivilege(c.Request.Context(), sqlc_queries.UpdateUserChatModelPrivilegeParams{
		ID:        int32(id),
		RateLimit: input.RateLimit,
		UpdatedBy: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrResourceNotFound("chat model privilege"))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to update user chat model privilege"))
		}
		return
	}
	output := ChatModelPrivilege{
		ID:            userChatModelPrivilege.ID,
		UserEmail:     input.UserEmail,
		ChatModelName: input.ChatModelName,
		RateLimit:     userChatModelPrivilege.RateLimit,
	}
	c.JSON(http.StatusOK, output)
}

func (h *UserChatModelPrivilegeHandler) DeleteUserChatModelPrivilege(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid user chat model privilege ID"))
		return
	}

	err = h.db.DeleteUserChatModelPrivilege(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrResourceNotFound("chat model privilege"))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to delete user chat model privilege"))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserChatModelPrivilegeHandler) UserChatModelPrivilegeByUserAndModelID(c *gin.Context) {
	_, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	var input struct {
		UserID      int32
		ChatModelID int32
	}
	err = c.ShouldBindJSON(&input)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	userChatModelPrivilege, err := h.db.UserChatModelPrivilegeByUserAndModelID(c.Request.Context(),
		sqlc_queries.UserChatModelPrivilegeByUserAndModelIDParams{
			UserID:      input.UserID,
			ChatModelID: input.ChatModelID,
		})

	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to get user chat model privilege"))
		return
	}

	c.JSON(http.StatusOK, userChatModelPrivilege)
}

func (h *UserChatModelPrivilegeHandler) ListUserChatModelPrivilegesByUserID(c *gin.Context) {
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	privileges, err := h.db.ListUserChatModelPrivilegesByUserID(c.Request.Context(), int32(userID))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to list privileges for user"))
		return
	}

	c.JSON(http.StatusOK, privileges)
}
