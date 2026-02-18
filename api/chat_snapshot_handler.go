package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatSnapshotHandler handles requests related to chat snapshots
type ChatSnapshotHandler struct {
	service *ChatSnapshotService
}

// NewChatSnapshotHandler creates a new handler instance
func NewChatSnapshotHandler(sqlc_q *sqlc_queries.Queries) *ChatSnapshotHandler {
	return &ChatSnapshotHandler{
		service: NewChatSnapshotService(sqlc_q),
	}
}

func (h *ChatSnapshotHandler) Register(router *gin.RouterGroup) {
	router.GET("/uuid/chat_snapshot/all", h.ChatSnapshotMetaByUserID)
	router.GET("/uuid/chat_snapshot/:uuid", h.GetChatSnapshot)
	router.POST("/uuid/chat_snapshot/:uuid", h.CreateChatSnapshot)
	router.PUT("/uuid/chat_snapshot/:uuid", h.UpdateChatSnapshotMetaByUUID)
	router.DELETE("/uuid/chat_snapshot/:uuid", h.DeleteChatSnapshot)
	router.GET("/uuid/chat_snapshot_search", h.ChatSnapshotSearch)
	router.POST("/uuid/chat_bot/:uuid", h.CreateChatBot)
}

func (h *ChatSnapshotHandler) CreateChatSnapshot(c *gin.Context) {
	chatSessionUuid := c.Param("uuid")
	user_id, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	uuid, err := h.service.CreateChatSnapshot(c.Request.Context(), chatSessionUuid, user_id)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to create chat snapshot")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, map[string]interface{}{
		"uuid": uuid,
	})

}

func (h *ChatSnapshotHandler) CreateChatBot(c *gin.Context) {
	chatSessionUuid := c.Param("uuid")
	user_id, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	uuid, err := h.service.CreateChatBot(c.Request.Context(), chatSessionUuid, user_id)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat bot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, map[string]interface{}{
		"uuid": uuid,
	})

}

func (h *ChatSnapshotHandler) GetChatSnapshot(c *gin.Context) {
	uuidStr := c.Param("uuid")
	snapshot, err := h.service.q.ChatSnapshotByUUID(c.Request.Context(), uuidStr)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get chat snapshot")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, snapshot)

}

func (h *ChatSnapshotHandler) ChatSnapshotMetaByUserID(c *gin.Context) {
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	// get type from query
	typ := c.Query("type")

	// Parse pagination parameters
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := int32(1)    // Default to page 1
	pageSize := int32(20) // Default to 20 items per page

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = int32(parsedPage)
		}
	}

	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err == nil && parsedPageSize > 0 && parsedPageSize <= 100 {
			pageSize = int32(parsedPageSize)
		}
	}

	offset := (page - 1) * pageSize

	chatSnapshots, err := h.service.q.ChatSnapshotMetaByUserID(c.Request.Context(), sqlc_queries.ChatSnapshotMetaByUserIDParams{
		UserID: userID,
		Typ:    typ,
		Limit:  pageSize,
		Offset: offset,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve chat snapshots"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Get total count for pagination
	totalCount, err := h.service.q.ChatSnapshotCountByUserIDAndType(c.Request.Context(), sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams{
		UserID: userID,
		Column2: typ,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve snapshot count"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	c.JSON(200, map[string]interface{}{
		"data":       chatSnapshots,
		"page":       page,
		"page_size":  pageSize,
		"total":      totalCount,
	})
}
func (h *ChatSnapshotHandler) UpdateChatSnapshotMetaByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	}
	err := c.ShouldBindJSON(&input)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	err = h.service.q.UpdateChatSnapshotMetaByUUID(c.Request.Context(), sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{
		Uuid:    uuid,
		Title:   input.Title,
		Summary: input.Summary,
		UserID:  userID,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat snapshot metadata"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(c.Request.Context(), uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, snapshot)

}

func (h *ChatSnapshotHandler) DeleteChatSnapshot(c *gin.Context) {
	uuid := c.Param("uuid")
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	_, err = h.service.q.DeleteChatSnapshot(c.Request.Context(), sqlc_queries.DeleteChatSnapshotParams{
		Uuid:   uuid,
		UserID: userID,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat snapshot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

}

func (h *ChatSnapshotHandler) ChatSnapshotSearch(c *gin.Context) {
	search := c.Query("search")
	if search == "" {
		var emptySlice []any // create an empty slice of integers
		c.JSON(200, emptySlice)
		return
	}
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	chatSnapshots, err := h.service.q.ChatSnapshotSearch(c.Request.Context(), sqlc_queries.ChatSnapshotSearchParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to search chat snapshots"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	c.JSON(200, chatSnapshots)
}
