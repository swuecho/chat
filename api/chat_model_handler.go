package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/pkg/errors"
	"github.com/swuecho/chat_backend/repository"
	"github.com/swuecho/chat_backend/service"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelHandler struct {
	serviceManager service.ServiceManager
}

func NewChatModelHandler(db *sqlc_queries.Queries) *ChatModelHandler {
	// Initialize repository and service layers
	repositoryManager := repository.NewCoreRepositoryManager(db)
	serviceManager := service.NewServiceManager(repositoryManager)
	
	return &ChatModelHandler{
		serviceManager: serviceManager,
	}
}

func (h *ChatModelHandler) Register(r *mux.Router) {
	// Chat model endpoints
	r.HandleFunc("/chat_model", h.ListSystemChatModels).Methods("GET")
	r.HandleFunc("/chat_model/default", h.GetDefaultChatModel).Methods("GET")
	r.HandleFunc("/chat_model/{id}", h.ChatModelByID).Methods("GET")
	r.HandleFunc("/chat_model", h.CreateChatModel).Methods("POST")
	r.HandleFunc("/chat_model/{id}", h.UpdateChatModel).Methods("PUT")
	r.HandleFunc("/chat_model/{id}", h.DeleteChatModel).Methods("DELETE")
}

func (h *ChatModelHandler) ListSystemChatModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Use service layer to get models with usage statistics
	modelsWithUsage, err := h.serviceManager.Model().GetSystemModelsWithUsage(ctx, "30 days")
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(modelsWithUsage)
}

func (h *ChatModelHandler) GetDefaultChatModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	model, err := h.serviceManager.Model().GetDefaultModel(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model)
}

func (h *ChatModelHandler) ChatModelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	if idStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("Model ID is required"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid model ID format"))
		return
	}

	ctx := r.Context()
	model, err := h.serviceManager.Model().GetModelByID(ctx, int32(id))
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model)
}

func (h *ChatModelHandler) CreateChatModel(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}

	// Parse request body
	var req service.ChatModelCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid request format"))
		return
	}

	// Set user ID from authentication
	req.UserID = userID

	// Validate required fields
	if req.Name == "" {
		errors.WriteErrorResponse(w, errors.ValidationFailed("name", "Model name is required"))
		return
	}
	if req.Label == "" {
		errors.WriteErrorResponse(w, errors.ValidationFailed("label", "Model label is required"))
		return
	}

	// Create model using service layer
	model, err := h.serviceManager.Model().CreateModel(ctx, req)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model)
}

func (h *ChatModelHandler) UpdateChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	if idStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("Model ID is required"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid model ID format"))
		return
	}

	// Get user ID from authentication
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}

	// Parse request body
	var req service.ChatModelUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid request format"))
		return
	}

	// TODO: Add authorization check to ensure user can update this model
	_ = userID

	// Update model using service layer
	model, err := h.serviceManager.Model().UpdateModel(ctx, int32(id), req)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model)
}

func (h *ChatModelHandler) DeleteChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	if idStr == "" {
		errors.WriteErrorResponse(w, errors.InvalidInput("Model ID is required"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.WriteErrorResponse(w, errors.InvalidInput("Invalid model ID format"))
		return
	}

	// Get user ID from authentication
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		errors.WriteErrorResponse(w, errors.Unauthorized("Authentication required"))
		return
	}

	// Delete model using service layer with authorization
	err = h.serviceManager.Model().DeleteModel(ctx, int32(id), userID)
	if err != nil {
		errors.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Chat model deleted successfully",
		"id":      id,
	})
}