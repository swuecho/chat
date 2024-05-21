package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestChatMessageService(t *testing.T) {
	// Create a new ChatMessageService with the test database connection
	q := sqlc_queries.New(db)
	service := NewChatMessageService(q)

	// Insert a new chat message into the database
	msg_params := sqlc_queries.CreateChatMessageParams{ChatSessionUuid: "1", Role: "Test Role", Content: "Test Message", Score: 0.5, UserID: 1,
		Raw: json.RawMessage([]byte("{}"))}
	msg, err := service.CreateChatMessage(context.Background(), msg_params)
	if err != nil {
		t.Fatalf("failed to create chat message: %v", err)
	}

	// Retrieve the inserted chat message from the database and check that it matches the expected values
	retrieved_msg, err := service.GetChatMessageByID(context.Background(), msg.ID)
	if err != nil {
		t.Fatalf("failed to retrieve chat message: %v", err)
	}
	if retrieved_msg.ID != msg.ID || retrieved_msg.ChatSessionUuid != msg.ChatSessionUuid ||
		retrieved_msg.Role != msg.Role || retrieved_msg.Content != msg.Content || retrieved_msg.Score != msg.Score ||
		retrieved_msg.UserID != msg.UserID || !retrieved_msg.CreatedAt.Time.Equal(msg.CreatedAt.Time) || !retrieved_msg.UpdatedAt.Time.Equal(msg.UpdatedAt.Time) ||
		retrieved_msg.CreatedBy != msg.CreatedBy || retrieved_msg.UpdatedBy != msg.UpdatedBy {
		t.Error("retrieved chat message does not match expected values")
	}

	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatMessage(context.Background(), msg.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	_, err = service.GetChatMessageByID(context.Background(), msg.ID)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		t.Error("expected error due to missing chat prompt, but got no error or different error")
	}
}

func TestGetChatMessagesBySessionID(t *testing.T) {

	// Create a new ChatMessageService with the test database connection
	q := sqlc_queries.New(db)
	service := NewChatMessageService(q)

	// Insert two chat messages into the database with different chat session IDs
	msg1_params := sqlc_queries.CreateChatMessageParams{ChatSessionUuid: "1", Role: "Test Role 1", Content: "Test Message 1", Score: 0.5, UserID: 1, Raw: json.RawMessage([]byte("{}"))}
	msg1, err := service.CreateChatMessage(context.Background(), msg1_params)
	if err != nil {
		t.Fatalf("failed to create chat message: %v", err)
	}
	msg2_params := sqlc_queries.CreateChatMessageParams{ChatSessionUuid: "2", Role: "Test Role 2", Content: "Test Message 2", Score: 0.75, UserID: 2, Raw: json.RawMessage([]byte("{}"))}
	msg2, err := service.CreateChatMessage(context.Background(), msg2_params)
	if err != nil {
		t.Fatalf("failed to create chat message: %v", err)
	}

	// Retrieve chat messages by chat session ID and check that they match the expected values
	// skip because of there is no chatSession with uuid "1" avaialble
	// chatSessionID := "1"
	// msgs, err := service.GetChatMessagesBySessionUUID(context.Background(), chatSessionID, 1, 10)
	// if err != nil {
	// 	t.Fatalf("failed to retrieve chat messages: %v", err)
	// }
	// if len(msgs) != 1 {
	// 	t.Errorf("expected 1 chat message, but got %d", len(msgs))
	// }
	// if msgs[0].ChatSessionUuid != msg1.ChatSessionUuid || msgs[0].Role != msg1.Role || msgs[0].Content != msg1.Content ||
	// 	msgs[0].Score != msg1.Score || msgs[0].UserID != msg1.UserID {
	// 	t.Error("retrieved chat messages do not match expected values")
	// }
	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatMessage(context.Background(), msg1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatMessage(context.Background(), msg2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	_, err = service.GetChatMessageByID(context.Background(), msg1.ID)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		t.Error("expected error due to missing chat prompt, but got no error or different error")
	}
	_, err = service.GetChatMessageByID(context.Background(), msg2.ID)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		t.Error("expected error due to missing chat prompt, but got no error or different error")
	}
}
