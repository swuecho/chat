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

func (h *ChatFileHandler) Register(router *gin.RouterGroup) {
	router.POST("/upload", h.ReceiveFile)
	router.GET("/chat_file/:uuid/list", h.ChatFilesBySessionUUID)
	router.GET("/download/:id", h.DownloadFile)
	router.DELETE("/download/:id", h.DeleteFile)
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

func (h *ChatFileHandler) ReceiveFile(c *gin.Context) {
	// Parse multipart form with size limit
	if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput(fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize)))
		return
	}

	// Get session UUID
	sessionUUID := c.PostForm("session-uuid")
	if sessionUUID == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("missing session UUID"))
		return
	}

	// Get user ID
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to get uploaded file"))
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing uploaded file: %v", err)
		}
	}()

	// Validate file type and extension
	mimeType := header.Header.Get("Content-Type")
	if !isValidFileType(mimeType, header.Filename) {
		RespondWithAPIErrorGin(c, ErrChatFileInvalidType.WithMessage(
			fmt.Sprintf("unsupported file type: %s or invalid extension for type", mimeType)))
		return
	}

	log.Printf("Uploading file: %s (%s, %d bytes)",
		header.Filename, mimeType, header.Size)

	// Validate file size
	if header.Size > maxUploadSize {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput(fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize)))
		return
	}

	// Read file into buffer with size limit
	var buf bytes.Buffer
	limitedReader := &io.LimitedReader{R: file, N: maxUploadSize}
	if _, err := io.Copy(&buf, limitedReader); err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to read uploaded file"))
		return
	}

	// Check if we hit the size limit
	if limitedReader.N <= 0 {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput(
			fmt.Sprintf("file exceeds maximum size of %d bytes", maxUploadSize)))
		return
	}
	// Create chat file record
	chatFile, err := h.service.q.CreateChatFile(c.Request.Context(), sqlc_queries.CreateChatFileParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
		Name:            header.Filename,
		Data:            buf.Bytes(),
		MimeType:        mimeType,
	})

	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to create chat file record"))
		return
	}

	// Clean up buffer
	buf.Reset()

	// Return success response
	c.JSON(http.StatusCreated, map[string]string{
		"url":  fmt.Sprintf("/download/%d", chatFile.ID),
		"name": header.Filename,
		"type": mimeType,
		"size": fmt.Sprintf("%d", header.Size),
	})
}

func (h *ChatFileHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	fileIdInt, err := strconv.ParseInt(fileID, 10, 32)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid file ID"))
		return
	}

	file, err := h.service.q.GetChatFileByID(c.Request.Context(), int32(fileIdInt))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrChatFileNotFound.WithMessage(fmt.Sprintf("file ID %d not found", fileIdInt)))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to get chat file"))
		}
		return
	}

	// Set proper content type from stored mime type
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	c.Header("Content-Length", strconv.Itoa(len(file.Data)))
	c.Data(200, file.MimeType, file.Data)
}

func (h *ChatFileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	fileIdInt, _ := strconv.ParseInt(fileID, 10, 32)
	_, err := h.service.q.DeleteChatFile(c.Request.Context(), int32(fileIdInt))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to delete chat file"))
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatFileHandler) ChatFilesBySessionUUID(c *gin.Context) {
	sessionUUID := c.Param("uuid")
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}
	chatFiles, err := h.service.q.ListChatFilesBySessionUUID(c.Request.Context(), sqlc_queries.ListChatFilesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
	})
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to list chat files for session"))
		return
	}

	if len(chatFiles) == 0 {
		c.String(200, "[]")
		return
	}
	c.JSON(200, chatFiles)
}
