package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatPromptHandler struct {
	service *ChatPromptService
}

func NewChatPromptHandler(sqlc_q *sqlc_queries.Queries) *ChatPromptHandler {
	promptService := NewChatPromptService(sqlc_q)
	return &ChatPromptHandler{
		service: promptService,
	}
}

func (h *ChatPromptHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_prompts", h.CreateChatPrompt).Methods(http.MethodPost)
	router.HandleFunc("/chat_prompts/users", h.GetChatPromptsByUserID).Methods(http.MethodGet)
	router.HandleFunc("/chat_prompts/{id}", h.GetChatPromptByID).Methods(http.MethodGet)
	router.HandleFunc("/chat_prompts/{id}", h.UpdateChatPrompt).Methods(http.MethodPut)
	router.HandleFunc("/chat_prompts/{id}", h.DeleteChatPrompt).Methods(http.MethodDelete)
	router.HandleFunc("/chat_prompts", h.GetAllChatPrompts).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_prompts/{uuid}", h.DeleteChatPromptByUUID).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_prompts/{uuid}", h.UpdateChatPromptByUUID).Methods(http.MethodPut)
}

func (h *ChatPromptHandler) CreateChatPrompt(w http.ResponseWriter, r *http.Request) {
	var promptParams sqlc_queries.CreateChatPromptParams
	err := json.NewDecoder(r.Body).Decode(&promptParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	prompt, err := h.service.CreateChatPrompt(r.Context(), promptParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

func (h *ChatPromptHandler) GetChatPromptByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat prompt ID", http.StatusBadRequest)
		return
	}
	prompt, err := h.service.GetChatPromptByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

func (h *ChatPromptHandler) UpdateChatPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat prompt ID", http.StatusBadRequest)
		return
	}
	var promptParams sqlc_queries.UpdateChatPromptParams
	err = json.NewDecoder(r.Body).Decode(&promptParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	promptParams.ID = int32(id)
	prompt, err := h.service.UpdateChatPrompt(r.Context(), promptParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

func (h *ChatPromptHandler) DeleteChatPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat prompt ID", http.StatusBadRequest)
		return
	}
	err = h.service.DeleteChatPrompt(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatPromptHandler) GetAllChatPrompts(w http.ResponseWriter, r *http.Request) {
	prompts, err := h.service.GetAllChatPrompts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prompts)
}

func (h *ChatPromptHandler) GetChatPromptsByUserID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	prompts, err := h.service.GetChatPromptsByUserID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(prompts)
}

func (h *ChatPromptHandler) DeleteChatPromptByUUID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatPromptByUUID(r.Context(), idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatPromptHandler) UpdateChatPromptByUUID(w http.ResponseWriter, r *http.Request) {
	var simple_msg SimpleChatMessage
	err := json.NewDecoder(r.Body).Decode(&simple_msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	prompt, err := h.service.UpdateChatPromptByUUID(r.Context(), simple_msg.Uuid, simple_msg.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prompt)
}
