package main

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
	"strings"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatFileHandler struct {
	service *ChatFileService
}

func NewChatFileHandler(sqlc_q *sqlc_queries.Queries) *ChatFileHandler {
	return &ChatFileHandler{service: NewChatFileService(sqlc_q)}
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

	sessionUUID := r.Header.Get("X-Session-Uuid")
	if sessionUUID == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("missing X-Session-Uuid header"))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	mimeType := r.Header.Get("Content-Type")
	var data []byte
	var filename string

	if strings.HasPrefix(mimeType, "multipart/form-data") {
		reader, err := r.MultipartReader()
		if err != nil {
			RespondWithAPIError(w, ErrValidationInvalidInput("failed to parse multipart form").WithDebugInfo(err.Error()))
			return
		}
		part, err := reader.NextPart()
		if err != nil {
			RespondWithAPIError(w, ErrValidationInvalidInput("failed to read form part").WithDebugInfo(err.Error()))
			return
		}
		filename = part.FileName()
		mimeType = part.Header.Get("Content-Type")
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, part); err != nil {
			RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("failed to read file data").WithDebugInfo(err.Error()))
			return
		}
		data = buf.Bytes()
	} else {
		filename = r.Header.Get("X-File-Name")
		limitedReader := io.LimitReader(r.Body, maxUploadSize+1)
		var buf bytes.Buffer
		n, err := io.CopyN(&buf, limitedReader, maxUploadSize+1)
		if err != nil && err != io.EOF {
			RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("failed to read request body").WithDebugInfo(err.Error()))
			return
		}
		if n > maxUploadSize {
			RespondWithAPIError(w, ErrValidationInvalidInput(fmt.Sprintf("file exceeds maximum size of %d bytes", maxUploadSize)))
			return
		}
		data = buf.Bytes()
	}

	file, err := h.service.CreateChatUpload(r.Context(), sqlc_queries.CreateChatFileParams{
		ChatSessionUuid: sessionUUID, UserID: userID,
		Name: filename, Data: data, MimeType: mimeType,
	})
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to create chat file record"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"url": fmt.Sprintf("/download/%d", file.ID), "name": filename,
		"type": mimeType, "size": fmt.Sprintf("%d", len(data)),
	})
}

func (h *ChatFileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	fileIdInt, err := strconv.ParseInt(fileID, 10, 32)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid file ID"))
		return
	}

	file, err := h.service.GetChatFile(r.Context(), int32(fileIdInt))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIError(w, ErrChatFileNotFound.WithMessage(fmt.Sprintf("file ID %d not found", fileIdInt)))
		} else {
			RespondWithAPIError(w, WrapError(err, "failed to get chat file"))
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
		RespondWithAPIError(w, WrapError(err, "failed to delete chat file"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatFileHandler) ChatFilesBySessionUUID(w http.ResponseWriter, r *http.Request) {
	sessionUUID := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	files, err := h.service.ListChatFilesBySession(r.Context(), sessionUUID, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "failed to list chat files for session"))
		return
	}

	// Return only safe metadata (not raw file data)
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


