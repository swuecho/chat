package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatCommentHandler struct {
	service *svc.ChatCommentService
}

func NewChatCommentHandler(service *svc.ChatCommentService) *ChatCommentHandler {
	return &ChatCommentHandler{service: service}
}

type chatCommentResponse struct {
	ID              int32     `json:"id"`
	UUID            string    `json:"uuid"`
	ChatSessionUUID string    `json:"chatSessionUuid"`
	ChatMessageUUID string    `json:"chatMessageUuid"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	CreatedBy       int32     `json:"createdBy"`
	UpdatedBy       int32     `json:"updatedBy"`
}

type commentWithAuthorResponse struct {
	UUID            string    `json:"uuid"`
	ChatMessageUUID string    `json:"chatMessageUuid,omitempty"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
	AuthorUsername  string    `json:"authorUsername"`
	AuthorEmail     string    `json:"authorEmail"`
}

func (h *ChatCommentHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	security := apicontract.BearerAuth()
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodPost, Path: "/uuid/chat_sessions/{sessionUUID}/chat_messages/{messageUUID}/comments", OperationID: "createChatComment",
		Summary: "Create a chat comment", Tags: []string{"Comments"}, SuccessStatus: http.StatusCreated, Security: security,
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("sessionUUID"), apicontract.UUIDPathParameter("messageUUID")},
	}, h.createChatComment)
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/uuid/chat_sessions/{sessionUUID}/comments", OperationID: "listSessionComments",
		Summary: "List comments for a chat session", Tags: []string{"Comments"}, SuccessStatus: http.StatusOK, Security: security,
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("sessionUUID")},
	}, h.getCommentsBySessionUUID)
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/uuid/chat_messages/{messageUUID}/comments", OperationID: "listMessageComments",
		Summary: "List comments for a chat message", Tags: []string{"Comments"}, SuccessStatus: http.StatusOK, Security: security,
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("messageUUID")},
	}, h.getCommentsByMessageUUID)
}

func (h *ChatCommentHandler) createChatComment(r *http.Request, req createCommentRequest) (chatCommentResponse, error) {
	vars := mux.Vars(r)
	sessionUUID := vars["sessionUUID"]
	messageUUID := vars["messageUUID"]

	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatCommentResponse{}, err
	}

	comment, err := h.service.CreateChatComment(r.Context(), svc.CreateChatCommentInput{
		Uuid:            uuid.New().String(),
		ChatSessionUuid: sessionUUID,
		ChatMessageUuid: messageUUID,
		Content:         req.Content,
		CreatedBy:       userID,
	})
	if err != nil {
		return chatCommentResponse{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat comment")
	}
	return chatCommentResponse{ID: comment.ID, UUID: comment.Uuid, ChatSessionUUID: comment.ChatSessionUuid, ChatMessageUUID: comment.ChatMessageUuid, Content: comment.Content, CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt, CreatedBy: comment.CreatedBy, UpdatedBy: comment.UpdatedBy}, nil
}

func (h *ChatCommentHandler) getCommentsBySessionUUID(r *http.Request, _ apicontract.NoBody) ([]commentWithAuthorResponse, error) {
	sessionUUID := mux.Vars(r)["sessionUUID"]

	comments, err := h.service.GetCommentsBySessionUUID(r.Context(), sessionUUID)
	if err != nil {
		return nil, dto.WrapError(dto.MapDatabaseError(err), "Failed to get comments by session")
	}
	response := make([]commentWithAuthorResponse, len(comments))
	for i, comment := range comments {
		response[i] = commentWithAuthorResponse{UUID: comment.Uuid, ChatMessageUUID: comment.ChatMessageUuid, Content: comment.Content, CreatedAt: comment.CreatedAt, AuthorUsername: comment.AuthorUsername, AuthorEmail: comment.AuthorEmail}
	}
	return response, nil
}

func (h *ChatCommentHandler) getCommentsByMessageUUID(r *http.Request, _ apicontract.NoBody) ([]commentWithAuthorResponse, error) {
	messageUUID := mux.Vars(r)["messageUUID"]

	comments, err := h.service.GetCommentsByMessageUUID(r.Context(), messageUUID)
	if err != nil {
		return nil, dto.WrapError(dto.MapDatabaseError(err), "Failed to get comments by message")
	}
	response := make([]commentWithAuthorResponse, len(comments))
	for i, comment := range comments {
		response[i] = commentWithAuthorResponse{UUID: comment.Uuid, Content: comment.Content, CreatedAt: comment.CreatedAt, AuthorUsername: comment.AuthorUsername, AuthorEmail: comment.AuthorEmail}
	}
	return response, nil
}
