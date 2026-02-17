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

// GinRegister registers routes with Gin router
func (h *ChatSnapshotHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/uuid/chat_snapshot/all", h.GinChatSnapshotMetaByUserID)
	rg.GET("/uuid/chat_snapshot/:uuid", h.GinGetChatSnapshot)
	rg.POST("/uuid/chat_snapshot/:uuid", h.GinCreateChatSnapshot)
	rg.PUT("/uuid/chat_snapshot/:uuid", h.GinUpdateChatSnapshotMetaByUUID)
	rg.DELETE("/uuid/chat_snapshot/:uuid", h.GinDeleteChatSnapshot)
	rg.GET("/uuid/chat_snapshot_search", h.GinChatSnapshotSearch)
	rg.POST("/uuid/chat_bot/:uuid", h.GinCreateChatBot)
}

func (h *ChatSnapshotHandler) GinCreateChatSnapshot(c *gin.Context) {
	chatSessionUuid := c.Param("uuid")
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	uuid, err := h.service.CreateChatSnapshot(c.Request.Context(), chatSessionUuid, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to create chat snapshot").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"uuid": uuid})
}

func (h *ChatSnapshotHandler) GinCreateChatBot(c *gin.Context) {
	chatSessionUuid := c.Param("uuid")
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	uuid, err := h.service.CreateChatBot(c.Request.Context(), chatSessionUuid, userID)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat bot"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"uuid": uuid})
}

func (h *ChatSnapshotHandler) GinGetChatSnapshot(c *gin.Context) {
	uuidStr := c.Param("uuid")
	snapshot, err := h.service.q.ChatSnapshotByUUID(c.Request.Context(), uuidStr)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat snapshot").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func (h *ChatSnapshotHandler) GinChatSnapshotMetaByUserID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	typ := c.Query("type")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := int32(1)
	pageSize := int32(20)

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
		apiErr.GinResponse(c)
		return
	}

	totalCount, err := h.service.q.ChatSnapshotCountByUserIDAndType(c.Request.Context(), sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams{
		UserID:  userID,
		Column2: typ,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve snapshot count"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data":       chatSnapshots,
		"page":       page,
		"page_size":  pageSize,
		"total":      totalCount,
	})
}

func (h *ChatSnapshotHandler) GinUpdateChatSnapshotMetaByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	var input struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr.GinResponse(c)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(c.Request.Context(), uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func (h *ChatSnapshotHandler) GinDeleteChatSnapshot(c *gin.Context) {
	uuid := c.Param("uuid")
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr.GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatSnapshotHandler) GinChatSnapshotSearch(c *gin.Context) {
	search := c.Query("search")
	if search == "" {
		c.JSON(http.StatusOK, []any{})
		return
	}
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr.GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, chatSnapshots)
}
