package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatFileHandler struct {
	service *ChatFileService
}

func NewChatFileHandler(sqlc_q *sqlc_queries.Queries) *ChatFileHandler {
	ChatFileService := NewChatFileService(sqlc_q)
	return &ChatFileHandler{
		service: ChatFileService,
	}
}

// GinRegister is an alias for Register for consistency with other handlers
func (h *ChatFileHandler) GinRegister(rg *gin.RouterGroup) {
	h.Register(rg)
}

func (h *ChatFileHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/upload", h.GinReceiveFile)
	rg.GET("/chat_file/:uuid/list", h.GinChatFilesBySessionUUID)
	rg.GET("/download/:id", h.GinDownloadFile)
	rg.DELETE("/download/:id", h.GinDeleteFile)
}

const (
	maxUploadSize = 32 << 20 // 32 MB
)

var allowedTypes = map[string]string{
	"image/jpeg":       ".jpg",
	"image/png":        ".png",
	"application/pdf":  ".pdf",
	"text/plain":       ".txt",
	"application/json": ".json",
}

// isValidFileType checks if the file type is allowed and matches the extension
func isValidFileType(mimeType, fileName string) bool {
	// Get expected extension for mime type
	expectedExt, ok := allowedTypes[mimeType]
	if !ok {
		return false
	}

	// Check if file has the expected extension
	return strings.HasSuffix(strings.ToLower(fileName), expectedExt)
}

func (h *ChatFileHandler) GinReceiveFile(c *gin.Context) {
	sessionUUID := c.PostForm("session-uuid")
	if sessionUUID == "" {
		ErrValidationInvalidInput("missing session UUID").GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		WrapError(err, "failed to get uploaded file").GinResponse(c)
		return
	}

	mimeType := file.Header.Get("Content-Type")
	if !isValidFileType(mimeType, file.Filename) {
		ErrChatFileInvalidType.WithMessage(
			fmt.Sprintf("unsupported file type: %s or invalid extension for type", mimeType)).GinResponse(c)
		return
	}

	log.Printf("Uploading file: %s (%s, %d bytes)",
		file.Filename, mimeType, file.Size)

	if file.Size > maxUploadSize {
		ErrValidationInvalidInput(fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize)).GinResponse(c)
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		WrapError(err, "failed to open uploaded file").GinResponse(c)
		return
	}
	defer openedFile.Close()

	var buf bytes.Buffer
	limitedReader := &io.LimitedReader{R: openedFile, N: maxUploadSize}
	if _, err := io.Copy(&buf, limitedReader); err != nil {
		WrapError(err, "failed to read uploaded file").GinResponse(c)
		return
	}

	if limitedReader.N <= 0 {
		ErrValidationInvalidInput(
			fmt.Sprintf("file exceeds maximum size of %d bytes", maxUploadSize)).GinResponse(c)
		return
	}

	chatFile, err := h.service.q.CreateChatFile(c.Request.Context(), sqlc_queries.CreateChatFileParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
		Name:            file.Filename,
		Data:            buf.Bytes(),
		MimeType:        mimeType,
	})

	if err != nil {
		WrapError(err, "failed to create chat file record").GinResponse(c)
		return
	}

	buf.Reset()

	c.JSON(http.StatusCreated, map[string]string{
		"url":  fmt.Sprintf("/download/%d", chatFile.ID),
		"name": file.Filename,
		"type": mimeType,
		"size": fmt.Sprintf("%d", file.Size),
	})
}

func (h *ChatFileHandler) GinDownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	fileIdInt, err := strconv.ParseInt(fileID, 10, 32)
	if err != nil {
		ErrValidationInvalidInput("invalid file ID").GinResponse(c)
		return
	}

	file, err := h.service.q.GetChatFileByID(c.Request.Context(), int32(fileIdInt))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrChatFileNotFound.WithMessage(fmt.Sprintf("file ID %d not found", fileIdInt)).GinResponse(c)
		} else {
			WrapError(err, "failed to get chat file").GinResponse(c)
		}
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	c.Header("Content-Length", strconv.Itoa(len(file.Data)))
	// Use application/octet-stream as fallback since MimeType is not stored
	c.Data(http.StatusOK, "application/octet-stream", file.Data)
}

func (h *ChatFileHandler) GinDeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	fileIdInt, _ := strconv.ParseInt(fileID, 10, 32)
	_, err := h.service.q.DeleteChatFile(c.Request.Context(), int32(fileIdInt))
	if err != nil {
		WrapError(err, "failed to delete chat file").GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatFileHandler) GinChatFilesBySessionUUID(c *gin.Context) {
	sessionUUID := c.Param("uuid")
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}
	chatFiles, err := h.service.q.ListChatFilesBySessionUUID(c.Request.Context(), sqlc_queries.ListChatFilesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
	})
	if err != nil {
		WrapError(err, "failed to list chat files for session").GinResponse(c)
		return
	}

	if len(chatFiles) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	c.JSON(http.StatusOK, chatFiles)
}
