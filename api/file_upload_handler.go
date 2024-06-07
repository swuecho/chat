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
	router.HandleFunc("/download/{id}", h.DownloadFile).Methods(http.MethodGet)
	router.HandleFunc("/download/{id}", h.DeleteFile).Methods(http.MethodDelete)
}

func (h *ChatFileHandler) ReceiveFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // limit your max input length!
	var buf bytes.Buffer
	// get formData(session-uuid) from request

	// Get the session-uuid from the form data
	sessionUUID := r.FormValue("session-uuid")
	fmt.Println("Session UUID:", sessionUUID)

	// get user-id from request
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	fmt.Println("User ID:", userID)

	// in your case file would be fileupload
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	name := header.Filename
	fmt.Printf("File name: %s\n", name)
	// Copy the file data to my buffer
	io.Copy(&buf, file)
	// inser into chat_file
	// check the raw content
	// select encode(data, 'escape') as data from chat_file;
	chatFile, err := h.service.q.CreateChatFile(r.Context(), sqlc_queries.CreateChatFileParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
		Name:            name,
		Data:            buf.Bytes(),
	})

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	buf.Reset()
	// return file name, file id as json

	url := fmt.Sprintf("/download/%d", chatFile.ID)

	json.NewEncoder(w).Encode(map[string]string{"url": url})
	w.WriteHeader(http.StatusOK)
}

func (h *ChatFileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	fileIdInt, _ := strconv.ParseInt(fileID, 10, 32)
	file, err := h.service.q.GetChatFileByID(r.Context(), int32(fileIdInt))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	w.WriteHeader(http.StatusOK)
}



