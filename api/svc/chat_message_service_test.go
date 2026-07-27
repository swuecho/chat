package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestChatMessageService(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatMessageService(q)

	msg_params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:    "1",
		Uuid:               "test-uuid-1",
		Role:               "Test Role",
		Content:            "Test Message",
		ReasoningContent:   "",
		Model:              "test-model",
		TokenCount:         100,
		Score:              0.5,
		UserID:             1,
		CreatedBy:          1,
		UpdatedBy:          1,
		LlmSummary:         "",
		Raw:                json.RawMessage([]byte("{}")),
		Artifacts:          json.RawMessage([]byte("[]")),
		SuggestedQuestions: json.RawMessage([]byte("[]")),
	}
	msg, err := service.CreateChatMessage(context.Background(), msg_params)
	if err != nil {
		t.Fatalf("failed to create chat message: %v", err)
	}

	retrieved_msg, err := service.GetChatMessageByID(context.Background(), msg.ID)
	if err != nil {
		t.Fatalf("failed to retrieve chat message: %v", err)
	}
	if retrieved_msg.ID != msg.ID || retrieved_msg.ChatSessionUuid != msg.ChatSessionUuid ||
		retrieved_msg.Role != msg.Role || retrieved_msg.Content != msg.Content || retrieved_msg.Score != msg.Score ||
		retrieved_msg.UserID != msg.UserID || !retrieved_msg.CreatedAt.Equal(msg.CreatedAt) || !retrieved_msg.UpdatedAt.Equal(msg.UpdatedAt) ||
		retrieved_msg.CreatedBy != msg.CreatedBy || retrieved_msg.UpdatedBy != msg.UpdatedBy {
		t.Error("retrieved chat message does not match expected values")
	}

	if err := service.DeleteChatMessage(context.Background(), msg.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	_, err = service.GetChatMessageByID(context.Background(), msg.ID)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Error("expected error due to missing chat prompt, but got no error or different error")
	}
}

func TestCreateChatMessageIsIdempotentWithinSession(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatMessageService(q)
	params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:    "1",
		Uuid:               "idempotent-message-uuid",
		Role:               "user",
		Content:            "original content",
		ReasoningContent:   "",
		Model:              "test-model",
		UserID:             1,
		CreatedBy:          1,
		UpdatedBy:          1,
		Raw:                json.RawMessage([]byte("{}")),
		Artifacts:          json.RawMessage([]byte("[]")),
		SuggestedQuestions: json.RawMessage([]byte("[]")),
	}

	first, err := service.CreateChatMessage(context.Background(), params)
	if err != nil {
		t.Fatalf("first CreateChatMessage() error = %v", err)
	}
	params.Content = "content from a retry must not overwrite the original"
	second, err := service.CreateChatMessage(context.Background(), params)
	if err != nil {
		t.Fatalf("retried CreateChatMessage() error = %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("retry created a duplicate: first ID %d, second ID %d", first.ID, second.ID)
	}
	if second.Content != first.Content {
		t.Fatalf("retry overwrote original content: got %q, want %q", second.Content, first.Content)
	}

	if err := service.DeleteChatMessage(context.Background(), first.ID); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	recreated, err := service.CreateChatMessage(context.Background(), params)
	if err != nil {
		t.Fatalf("recreating a soft-deleted message error = %v", err)
	}
	if recreated.ID == first.ID {
		t.Fatalf("recreating a deleted UUID returned deleted row ID %d", recreated.ID)
	}
	if err := service.DeleteChatMessage(context.Background(), recreated.ID); err != nil {
		t.Fatalf("recreated message cleanup failed: %v", err)
	}
}

func TestGetChatMessagesBySessionID(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatMessageService(q)

	msg1_params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:    "1",
		Uuid:               "test-uuid-1",
		Role:               "Test Role 1",
		Content:            "Test Message 1",
		ReasoningContent:   "",
		Model:              "test-model",
		TokenCount:         100,
		Score:              0.5,
		UserID:             1,
		CreatedBy:          1,
		UpdatedBy:          1,
		LlmSummary:         "",
		Raw:                json.RawMessage([]byte("{}")),
		Artifacts:          json.RawMessage([]byte("[]")),
		SuggestedQuestions: json.RawMessage([]byte("[]")),
	}
	msg1, err := service.CreateChatMessage(context.Background(), msg1_params)
	if err != nil {
		t.Fatalf("failed to create chat message: %v", err)
	}
	msg2_params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:    "2",
		Uuid:               "test-uuid-2",
		Role:               "Test Role 2",
		Content:            "Test Message 2",
		ReasoningContent:   "",
		Model:              "test-model",
		TokenCount:         100,
		Score:              0.75,
		UserID:             2,
		CreatedBy:          2,
		UpdatedBy:          2,
		LlmSummary:         "",
		Raw:                json.RawMessage([]byte("{}")),
		Artifacts:          json.RawMessage([]byte("[]")),
		SuggestedQuestions: json.RawMessage([]byte("[]")),
	}
	msg2, err := service.CreateChatMessage(context.Background(), msg2_params)
	if err != nil {
		t.Fatalf("failed to create chat message: %v", err)
	}

	if err := service.DeleteChatMessage(context.Background(), msg1.ID); err != nil {
		t.Fatalf("failed to delete chat message: %v", err)
	}
	if err := service.DeleteChatMessage(context.Background(), msg2.ID); err != nil {
		t.Fatalf("failed to delete chat message: %v", err)
	}

	_, err = service.GetChatMessageByID(context.Background(), msg1.ID)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Error("expected error due to missing chat message, but got no error or different error")
	}
	_, err = service.GetChatMessageByID(context.Background(), msg2.ID)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Error("expected error due to missing chat message, but got no error or different error")
	}
}
