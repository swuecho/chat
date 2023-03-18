package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chatgpt_backend/auth"
	"github.com/swuecho/chatgpt_backend/sqlc_queries"
)

type ChatMessageHandler struct {
	service *ChatMessageService
}

func NewChatMessageHandler(service *ChatMessageService) *ChatMessageHandler {
	return &ChatMessageHandler{
		service: service,
	}
}

func (h *ChatMessageHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_messages", h.CreateChatMessage).Methods(http.MethodPost)
	router.HandleFunc("/chat_messages/{id}", h.GetChatMessageByID).Methods(http.MethodGet)
	router.HandleFunc("/chat_messages/{id}", h.UpdateChatMessage).Methods(http.MethodPut)
	router.HandleFunc("/chat_messages/{id}", h.DeleteChatMessage).Methods(http.MethodDelete)
	router.HandleFunc("/chat_messages", h.GetAllChatMessages).Methods(http.MethodGet)

	router.HandleFunc("/uuid/chat_messages/{uuid}", h.GetChatMessageByUUID).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_messages/{uuid}", h.UpdateChatMessageByUUID).Methods(http.MethodPut)
	router.HandleFunc("/uuid/chat_messages/{uuid}", h.DeleteChatMessageByUUID).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_messages/chat_sessions/{uuid}", h.GetChatHistoryBySessionUUID).Methods(http.MethodGet)

}

type userIdContextKey string

const userIDKey = userIdContextKey("userID")

func AuthenticateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := "" //getTokenFromHeader(r.Header)
		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "invalid auth token", http.StatusUnauthorized)
			return
		}

		// // Check if the user has permission to read/write messages in the chat
		// sessionID, err := getSessionIDFromRequest(r)
		// if err != nil {
		// 	http.Error(w, "invalid session ID", http.StatusBadRequest)
		// 	return
		// }
		// if !h.sessionService.HasSessionPermission(sessionID, userID) {
		// 	http.Error(w, "user does not have permission to view or send messages", http.StatusForbidden)
		// 	return
		// }

		// Pass on the authenticated user ID to the next handler
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *ChatMessageHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	var messageParams sqlc_queries.CreateChatMessageParams
	err := json.NewDecoder(r.Body).Decode(&messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message, err := h.service.CreateChatMessage(r.Context(), messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) GetChatMessageByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat message ID", http.StatusBadRequest)
		return
	}
	message, err := h.service.GetChatMessageByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) UpdateChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat message ID", http.StatusBadRequest)
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageParams
	err = json.NewDecoder(r.Body).Decode(&messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	messageParams.ID = int32(id)
	message, err := h.service.UpdateChatMessage(r.Context(), messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func (h *ChatMessageHandler) DeleteChatMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid chat message ID", http.StatusBadRequest)
		return
	}
	err = h.service.DeleteChatMessage(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatMessageHandler) GetAllChatMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.service.GetAllChatMessages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}

// GetChatMessageByUUID get chat message by uuid
func (h *ChatMessageHandler) GetChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	message, err := h.service.GetChatMessageByUUID(r.Context(), uuidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(message)
}

// UpdateChatMessageByUUID update chat message by uuid
func (h *ChatMessageHandler) UpdateChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	var simple_msg SimpleChatMessage
	err := json.NewDecoder(r.Body).Decode(&simple_msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println(simple_msg)
	var messageParams sqlc_queries.UpdateChatMessageByUUIDParams
	messageParams.Uuid = simple_msg.Uuid
	messageParams.Content = simple_msg.Text
	log.Println(messageParams)
	message, err := h.service.UpdateChatMessageByUUID(r.Context(), messageParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(message)
}

// DeleteChatMessageByUUID delete chat message by uuid
func (h *ChatMessageHandler) DeleteChatMessageByUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatMessageByUUID(r.Context(), uuidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetChatMessagesBySessionUUID get chat messages by session uuid
func (h *ChatMessageHandler) GetChatMessagesBySessionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 200
	}

	messages, err := h.service.GetChatMessagesBySessionUUID(r.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	simple_msgs := lo.Map(messages, func(message sqlc_queries.ChatMessage, _ int) SimpleChatMessage {
		return SimpleChatMessage{
			DateTime:  time.Now().Local().Format("2006-01-02 15:04:05"),
			Text:      message.Content,
			Inversion: message.Role != "user",
			Error:     false,
			Loading:   false,
		}
	})
	json.NewEncoder(w).Encode(simple_msgs)
}

// GetChatMessagesBySessionUUID get chat messages by session uuid
func (h *ChatMessageHandler) GetChatHistoryBySessionUUID(w http.ResponseWriter, r *http.Request) {
	uuidStr := mux.Vars(r)["uuid"]
	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 200
	}
	simple_msgs, err := h.service.GetChatHistoryBySessionUUID(r.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(simple_msgs)
}
