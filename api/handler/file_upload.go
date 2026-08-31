package handler

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
)

type ChatFileHandler struct {
	service *svc.ChatFileService
}

type uploadFileContractRequest struct {
	File        apicontract.BinaryBody `json:"file" jsonschema:"required,format=binary"`
	SessionUUID string                 `json:"session-uuid" jsonschema:"required,format=uuid"`
}

func NewChatFileHandler(service *svc.ChatFileService) *ChatFileHandler {
	return &ChatFileHandler{service: service}
}

func (h *ChatFileHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	router.HandleFunc("/upload", endpoint(h.ReceiveFile)).Methods(http.MethodPost)
	apicontract.DocumentJSON[uploadFileContractRequest, fileUploadHTTPResponse](registry, apicontract.Operation{
		Method: http.MethodPost, Path: "/upload", OperationID: "uploadChatFile", Summary: "Upload a chat file",
		Tags: []string{"Files"}, SuccessStatus: http.StatusCreated, Security: apicontract.BearerAuth(), RequestContentType: "multipart/form-data",
	})
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/chat_file/{uuid}/list", OperationID: "listChatFiles",
		Summary: "List files for a chat session", Tags: []string{"Files"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(),
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")},
	}, h.chatFilesBySessionUUID)
	router.HandleFunc("/download/{id}", endpoint(h.DownloadFile)).Methods(http.MethodGet)
	router.HandleFunc("/download/{id}", endpoint(h.DeleteFile)).Methods(http.MethodDelete)
	idParameter := []apicontract.Parameter{apicontract.PositiveIntegerPathParameter("id")}
	apicontract.DocumentJSON[apicontract.NoBody, apicontract.BinaryBody](registry, apicontract.Operation{Method: http.MethodGet, Path: "/download/{id}", OperationID: "downloadChatFile", Summary: "Download a chat file", Tags: []string{"Files"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(), Parameters: idParameter, ResponseContentType: "application/octet-stream"})
	apicontract.DocumentJSON[apicontract.NoBody, apicontract.NoBody](registry, apicontract.Operation{Method: http.MethodDelete, Path: "/download/{id}", OperationID: "deleteChatFile", Summary: "Delete a chat file", Tags: []string{"Files"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(), Parameters: idParameter})
}

const maxUploadSize = 32 << 20 // 32MB

func (h *ChatFileHandler) ReceiveFile(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return dto.ErrValidationInvalidInput(fmt.Sprintf("file too large, max size is %d bytes", maxUploadSize))
	}

	sessionUUID := r.FormValue("session-uuid")
	if err := validation.UUID("session-uuid", sessionUUID, true); err != nil {
		return dto.ErrValidationInvalidInput(err.Error())
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return dto.ErrValidationInvalidInput("failed to read uploaded file").WithDebugInfo(err.Error())
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return dto.ErrInternalUnexpected.WithDetail("failed to read file data").WithDebugInfo(err.Error())
	}

	chatFile, err := h.service.CreateChatUpload(r.Context(), svc.CreateChatFileInput{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
		Name:            header.Filename,
		Data:            buf.Bytes(),
		MimeType:        mimeType,
	})
	if err != nil {
		return dto.WrapError(err, "failed to create chat file record")
	}
	return respondJSON(w, http.StatusCreated, fileUploadHTTPResponse{URL: fmt.Sprintf("/download/%d", chatFile.ID),
		Name: header.Filename, Type: mimeType, Size: fmt.Sprintf("%d", header.Size)})
}

func (h *ChatFileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) error {
	fileID, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}

	file, err := h.service.GetChatFile(r.Context(), fileID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Length", strconv.Itoa(len(file.Data)))
	if _, err := w.Write(file.Data); err != nil {
		slog.Info("Failed to write file data", "error", err)
	}
	return nil
}

func (h *ChatFileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	fileID, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}

	// Verify file ownership before deletion
	file, err := h.service.GetChatFile(r.Context(), fileID)
	if err != nil {
		return dto.WrapError(err, "failed to get chat file")
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	if file.UserID != userID {
		return dto.ErrAuthAccessDenied.WithMessage("You do not own this file")
	}

	if err := h.service.DeleteChatFile(r.Context(), fileID); err != nil {
		return dto.WrapError(err, "failed to delete chat file")
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatFileHandler) chatFilesBySessionUUID(r *http.Request, _ apicontract.NoBody) ([]fileMetaHTTPResponse, error) {
	sessionUUID := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return nil, err
	}

	files, err := h.service.ListChatFilesBySession(r.Context(), sessionUUID, userID)
	if err != nil {
		return nil, dto.WrapError(err, "failed to list chat files for session")
	}
	meta := make([]fileMetaHTTPResponse, 0, len(files))
	for _, f := range files {
		meta = append(meta, fileMetaHTTPResponse{ID: f.ID, Name: f.Name})
	}
	return meta, nil
}
