package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

func (h *ChatFileHandler) Register(router *mux.Router) {
	router.HandleFunc("/upload", h.ReceiveFile).Methods(http.MethodPost)
	router.HandleFunc("/chat_file/{uuid}/list", h.ChatFilesBySessionUUID).Methods(http.MethodGet)
	router.HandleFunc("/download/{id}", h.DownloadFile).Methods(http.MethodGet)
	router.HandleFunc("/download/{id}", h.DeleteFile).Methods(http.MethodDelete)
}

const (
	maxUploadSize = 32 << 20 // 32 MB
	allowedTypes  = "image/jpeg,image/png,application/pdf,text/plain"
)

func (h *ChatFileHandler) ReceiveFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with size limit
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput(fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize)))
		return
	}

	// Get session UUID
	sessionUUID := r.FormValue("session-uuid")
	if sessionUUID == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("missing session UUID"))
		return
	}

	// Get user ID
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get uploaded file"))
		return
	}
	defer file.Close()

	// Validate file type
	mimeType := header.Header.Get("Content-Type")
	if !strings.Contains(allowedTypes, mimeType) {
		RespondWithAPIError(w, ErrValidationInvalidInput(fmt.Sprintf("unsupported file type: %s", mimeType)))
		return
	}

	// Validate file size
	if header.Size > maxUploadSize {
		RespondWithAPIError(w, ErrValidationInvalidInput(fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize)))
		return
	}

	// Read file into buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to read uploaded file"))
		return
	}
	fmt.Printf("File name: %s\n", name)
	// Copy the file data to my buffer
	io.Copy(&buf, file)
	// Create chat file record
	chatFile, err := h.service.q.CreateChatFile(r.Context(), sqlc_queries.CreateChatFileParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
		Name:            header.Filename,
		Data:            buf.Bytes(),
		MimeType:        mimeType,
	})

	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to create chat file record"))
		return
	}

	// Clean up buffer
	buf.Reset()

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"url":  fmt.Sprintf("/download/%d", chatFile.ID),
		"name": header.Filename,
		"type": mimeType,
		"size": fmt.Sprintf("%d", header.Size),
	})
}

func (h *ChatFileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	fileIdInt, _ := strconv.ParseInt(fileID, 10, 32)
	file, err := h.service.q.GetChatFileByID(r.Context(), int32(fileIdInt))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to get chat file"))
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	// w.Header().Set("Content-Disposition",fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Write(file.Data)
}

func (h *ChatFileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	fileIdInt, _ := strconv.ParseInt(fileID, 10, 32)
	_, err := h.service.q.DeleteChatFile(r.Context(), int32(fileIdInt))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to delete chat file"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatFileHandler) ChatFilesBySessionUUID(w http.ResponseWriter, r *http.Request) {
	sessionUUID := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDetail("missing or invalid user ID"))
		return
	}
	chatFiles, err := h.service.q.ListChatFilesBySessionUUID(r.Context(), sqlc_queries.ListChatFilesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
	})
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to list chat files for session"))
		return
	}
	w.WriteHeader(http.StatusOK)

	if len(chatFiles) == 0 {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(chatFiles)
}
