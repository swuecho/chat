package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type ChatFileHandler struct {
	service *svc.ChatFileService
}

func NewChatFileHandler(sqlc_q *sqlc_queries.Queries) *ChatFileHandler {
	return &ChatFileHandler{service: svc.NewChatFileService(sqlc_q)}
}

func (h *ChatFileHandler) Register(router *mux.Router) {
	router.HandleFunc("/upload", h.ReceiveFile).Methods(http.MethodPost)
	router.HandleFunc("/chat_file/{uuid}/list", h.ChatFilesBySessionUUID).Methods(http.MethodGet)
	router.HandleFunc("/download/{id}", h.DownloadFile).Methods(http.MethodGet)
	router.HandleFunc("/download/{id}", h.DeleteFile).Methods(http.MethodDelete)
}

const maxUploadSize = 32 << 20 // 32MB

func (h *ChatFileHandler) ReceiveFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput(
			fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize)))
		return
	}

	sessionUUID := r.FormValue("session-uuid")
	if sessionUUID == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("missing session UUID"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("failed to read uploaded file").WithDebugInfo(err.Error()))
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("failed to read file data").WithDebugInfo(err.Error()))
		return
	}

	chatFile, err := h.service.CreateChatUpload(r.Context(), sqlc_queries.CreateChatFileParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
		Name:            header.Filename,
		Data:            buf.Bytes(),
		MimeType:        mimeType,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to create chat file record"))
		return
	}

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
	fileIdInt, err := strconv.ParseInt(fileID, 10, 32)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid file ID"))
		return
	}

	file, err := h.service.GetChatFile(r.Context(), int32(fileIdInt))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.ErrChatFileNotFound.WithMessage(fmt.Sprintf("file ID %d not found", fileIdInt)))
		} else {
			dto.RespondWithAPIError(w, dto.WrapError(err, "failed to get chat file"))
		}
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Length", strconv.Itoa(len(file.Data)))
	if _, err := w.Write(file.Data); err != nil {
		log.Printf("Failed to write file data: %v", err)
	}
}

func (h *ChatFileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	fileIdInt, _ := strconv.ParseInt(fileID, 10, 32)
	if err := h.service.DeleteChatFile(r.Context(), int32(fileIdInt)); err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to delete chat file"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatFileHandler) ChatFilesBySessionUUID(w http.ResponseWriter, r *http.Request) {
	sessionUUID := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	files, err := h.service.ListChatFilesBySession(r.Context(), sessionUUID, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "failed to list chat files for session"))
		return
	}

	type fileMeta struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	}
	meta := make([]fileMeta, 0, len(files))
	for _, f := range files {
		meta = append(meta, fileMeta{ID: f.ID, Name: f.Name})
	}
	json.NewEncoder(w).Encode(meta)
}
