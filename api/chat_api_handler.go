package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatAPIHandler struct {
	db *sqlc_queries.Queries
}

func NewChatAPIHandler(db *sqlc_queries.Queries) *ChatAPIHandler {
	return &ChatAPIHandler{
		db: db,
	}
}

func (h *ChatAPIHandler) Register(r *mux.Router) {

	// Assuming db is an instance of the SQLC generated DB struct
	//handler := NewChatAPIHandler(db)
	// r := mux.NewRouter()

	r.HandleFunc("/chat_apis", h.ListChatAPIs).Methods("GET")
	r.HandleFunc("/chat_apis/{id}", h.ChatAPIByID).Methods("GET")
	r.HandleFunc("/chat_apis", h.CreateChatAPI).Methods("POST")
	r.HandleFunc("/chat_apis/{id}", h.UpdateChatAPI).Methods("PUT")
	r.HandleFunc("/chat_apis/{id}", h.DeleteChatAPI).Methods("DELETE")
	r.HandleFunc("/chat_apis/default", h.GetDefaultChatAPI).Methods("GET")
}

func (h *ChatAPIHandler) ListChatAPIs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chatAPIs, err := h.db.ListChatAPIs(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error listing chat APIs: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatAPIs)
}

func (h *ChatAPIHandler) ChatAPIByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid chat API ID"))
		return
	}

	chatAPI, err := h.db.ChatAPIByID(ctx, int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error retrieving chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatAPI)
}

func (h *ChatAPIHandler) CreateChatAPI(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name          string `json:"name"`
		Label         string `json:"label"`
		IsDefault     bool   `json:"is_default"`
		URL           string `json:"url"`
		APIAuthHeader string `json:"api_auth_header"`
		APIAuthKey    string `json:"api_auth_key"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request body"))
		return
	}

	chatAPI, err := h.db.CreateChatAPI(r.Context(), sqlc_queries.CreateChatAPIParams{
		Name:          input.Name,
		Label:         input.Label,
		IsDefault:     input.IsDefault,
		Url:           input.URL,
		ApiAuthHeader: input.APIAuthHeader,
		ApiAuthKey:    input.APIAuthKey,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chatAPI)
}

func (h *ChatAPIHandler) UpdateChatAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid chat API ID"))
		return
	}

	var input struct {
		Name          string `json:"name"`
		Label         string `json:"label"`
		IsDefault     bool   `json:"is_default"`
		URL           string `json:"url"`
		APIAuthHeader string `json:"api_auth_header"`
		APIAuthKey    string `json:"api_auth_key"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request body"))
		return
	}

	chatAPI, err := h.db.UpdateChatAPI(r.Context(), sqlc_queries.UpdateChatAPIParams{
		ID:            int32(id),
		Name:          input.Name,
		Label:         input.Label,
		IsDefault:     input.IsDefault,
		Url:           input.URL,
		ApiAuthHeader: input.APIAuthHeader,
		ApiAuthKey:    input.APIAuthKey,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error updating chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatAPI)
}

func (h *ChatAPIHandler) DeleteChatAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid chat API ID"))
		return
	}

	err = h.db.DeleteChatAPI(r.Context(), int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error deleting chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ChatAPIHandler) GetDefaultChatAPI(w http.ResponseWriter, r *http.Request) {
	chatAPI, err := h.db.GetDefaultChatAPI(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error retrieving default chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatAPI)
}
