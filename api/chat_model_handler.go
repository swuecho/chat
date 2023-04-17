package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelHandler struct {
	db *sqlc_queries.Queries
}

func NewChatModelHandler(db *sqlc_queries.Queries) *ChatModelHandler {
	return &ChatModelHandler{
		db: db,
	}
}

func (h *ChatModelHandler) Register(r *mux.Router) {

	// Assuming db is an instance of the SQLC generated DB struct
	//handler := NewChatModelHandler(db)
	// r := mux.NewRouter()

	r.HandleFunc("/chat_model", h.ListChatModels).Methods("GET")
	r.HandleFunc("/chat_model/default", h.GetDefaultChatModel).Methods("GET")
	r.HandleFunc("/chat_model/{id}", h.ChatModelByID).Methods("GET")
	r.HandleFunc("/chat_model", h.CreateChatModel).Methods("POST")
	r.HandleFunc("/chat_model/{id}", h.UpdateChatModel).Methods("PUT")
	r.HandleFunc("/chat_model/{id}", h.DeleteChatModel).Methods("DELETE")
}

func (h *ChatModelHandler) ListChatModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ChatModels, err := h.db.ListChatModels(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error listing chat APIs: %s", err.Error())))
		return
	}


	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatModels)
}

func (h *ChatModelHandler) ChatModelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid chat API ID"))
		return
	}

	ChatModel, err := h.db.ChatModelByID(ctx, int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error retrieving chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatModel)
}

func (h *ChatModelHandler) CreateChatModel(w http.ResponseWriter, r *http.Request) {
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

	ChatModel, err := h.db.CreateChatModel(r.Context(), sqlc_queries.CreateChatModelParams{
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
	json.NewEncoder(w).Encode(ChatModel)
}

func (h *ChatModelHandler) UpdateChatModel(w http.ResponseWriter, r *http.Request) {
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

	ChatModel, err := h.db.UpdateChatModel(r.Context(), sqlc_queries.UpdateChatModelParams{
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
	json.NewEncoder(w).Encode(ChatModel)
}


func (h *ChatModelHandler) DeleteChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid chat API ID"))
		return
	}

	err = h.db.DeleteChatModel(r.Context(), int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error deleting chat API: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ChatModelHandler) GetDefaultChatModel(w http.ResponseWriter, r *http.Request) {
	ChatModel, err := h.db.GetDefaultChatModel(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error retrieving default chat API: %s", err.Error())))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatModel)
}
