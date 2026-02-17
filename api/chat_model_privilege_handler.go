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

// GinRegister registers routes with Gin router (note: these go under /admin)
func (h *UserChatModelPrivilegeHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/user_chat_model_privilege", h.GinListUserChatModelPrivileges)
	rg.POST("/user_chat_model_privilege", h.GinCreateUserChatModelPrivilege)
	rg.DELETE("/user_chat_model_privilege/:id", h.GinDeleteUserChatModelPrivilege)
	rg.PUT("/user_chat_model_privilege/:id", h.GinUpdateUserChatModelPrivilege)
}

type ChatModelPrivilege struct {
	ID            int32  `json:"id"`
	FullName      string `json:"fullName"`
	UserEmail     string `json:"userEmail"`
	ChatModelName string `json:"chatModelName"`
	RateLimit     int32  `json:"rateLimit"`
}

// =============================================================================
// Gin Handlers
// =============================================================================

func (h *UserChatModelPrivilegeHandler) GinListUserChatModelPrivileges(c *gin.Context) {
	userChatModelRows, err := h.db.ListUserChatModelPrivilegesRateLimit(c.Request.Context())
	if err != nil {
		WrapError(err, "failed to list user chat model privileges").GinResponse(c)
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

func (h *UserChatModelPrivilegeHandler) GinCreateUserChatModelPrivilege(c *gin.Context) {
	var input ChatModelPrivilege
	if err := c.ShouldBindJSON(&input); err != nil {
		ErrValidationInvalidInput("failed to parse request body").GinResponse(c)
		return
	}

	if input.UserEmail == "" {
		ErrValidationInvalidInput("user email is required").GinResponse(c)
		return
	}
	if input.ChatModelName == "" {
		ErrValidationInvalidInput("chat model name is required").GinResponse(c)
		return
	}
	if input.RateLimit <= 0 {
		ErrValidationInvalidInput("rate limit must be positive").WithMessage(
			fmt.Sprintf("invalid rate limit: %d", input.RateLimit)).GinResponse(c)
		return
	}

	log.Printf("Creating chat model privilege for user %s with model %s",
		input.UserEmail, input.ChatModelName)

	user, err := h.db.GetAuthUserByEmail(c.Request.Context(), input.UserEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrResourceNotFound("user").WithMessage(
				fmt.Sprintf("user with email %s not found", input.UserEmail)).GinResponse(c)
		} else {
			WrapError(err, "failed to get user by email").GinResponse(c)
		}
		return
	}

	chatModel, err := h.db.ChatModelByName(c.Request.Context(), input.ChatModelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrChatModelNotFound.WithMessage(fmt.Sprintf("chat model %s not found", input.ChatModelName)).GinResponse(c)
		} else {
			WrapError(err, "failed to get chat model").GinResponse(c)
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
			ErrResourceNotFound("chat model privilege").GinResponse(c)
		} else {
			WrapError(err, "failed to create user chat model privilege").GinResponse(c)
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

func (h *UserChatModelPrivilegeHandler) GinUpdateUserChatModelPrivilege(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid user chat model privilege ID").GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	var input ChatModelPrivilege
	if err := c.ShouldBindJSON(&input); err != nil {
		ErrValidationInvalidInput("failed to parse request body").GinResponse(c)
		return
	}

	if input.RateLimit <= 0 {
		ErrValidationInvalidInput("rate limit must be positive").GinResponse(c)
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
			ErrResourceNotFound("chat model privilege").GinResponse(c)
		} else {
			WrapError(err, "failed to update user chat model privilege").GinResponse(c)
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

func (h *UserChatModelPrivilegeHandler) GinDeleteUserChatModelPrivilege(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid user chat model privilege ID").GinResponse(c)
		return
	}

	err = h.db.DeleteUserChatModelPrivilege(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrResourceNotFound("chat model privilege").GinResponse(c)
		} else {
			WrapError(err, "failed to delete user chat model privilege").GinResponse(c)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
