package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/swuecho/chatgpt_backend/sqlc_queries"
)

func TestChatPromptService(t *testing.T) {
	// Create a new ChatPromptService with the test database connection
	q := sqlc_queries.New(db)
	service := NewChatPromptService(q)

	// Insert a new chat prompt into the database
	prompt_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic", Role: "Test Role", Content: "Test Content", UserID: 1}
	prompt, err := service.CreateChatPrompt(context.Background(), prompt_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	// Retrieve the chat prompt by ID and check that it matches the expected values
	retrievedPrompt, err := service.GetChatPromptByID(context.Background(), prompt.ID)
	if err != nil {
		t.Fatalf("failed to get chat prompt: %v", err)
	}
	if retrievedPrompt.ChatSessionUuid != prompt.ChatSessionUuid || retrievedPrompt.Role != prompt.Role || retrievedPrompt.Content != prompt.Content || retrievedPrompt.Score != prompt.Score || retrievedPrompt.UserID != prompt.UserID {
		t.Error("retrieved chat prompt does not match expected values")
	}

	// Update the chat prompt and check that it was updated in the database
	updated_params := sqlc_queries.UpdateChatPromptParams{ID: prompt.ID, ChatSessionUuid: "Updated Test Topic", Role: "Updated Test Role", Content: "Updated Test Content", Score: 0.75}
	if _, err := service.UpdateChatPrompt(context.Background(), updated_params); err != nil {
		t.Fatalf("failed to update chat prompt: %v", err)
	}
	retrievedPrompt, err = service.GetChatPromptByID(context.Background(), prompt.ID)
	if err != nil {
		t.Fatalf("failed to get chat prompt: %v", err)
	}
	if retrievedPrompt.ChatSessionUuid != updated_params.ChatSessionUuid || retrievedPrompt.Role != updated_params.Role || retrievedPrompt.Content != updated_params.Content || retrievedPrompt.Score != updated_params.Score {
		t.Error("retrieved chat prompt does not match expected values")
	}

	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatPrompt(context.Background(), prompt.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	_, err = service.GetChatPromptByID(context.Background(), prompt.ID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		print("Chat prompt deleted successfully")
	}
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Error("expected error due to missing chat prompt, but got no error or different error")
	}

	_, err = service.q.GetChatPromptsBySessionUUID(context.Background(), "12324")

	if err != nil {
		t.Error("expected error due to missing chat prompt, but got no error or different error")
	}
}

func TestGetAllChatPrompts(t *testing.T) {
	q := sqlc_queries.New(db)
	service := NewChatPromptService(q)

	// Insert two chat prompts into the database
	prompt1_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic 1", Role: "Test Role 1", Content: "Test Content 1", UserID: 1}
	prompt1, err := service.CreateChatPrompt(context.Background(), prompt1_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}
	prompt2_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic 2", Role: "Test Role 2", Content: "Test Content 2", UserID: 2}
	prompt2, err := service.CreateChatPrompt(context.Background(), prompt2_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	// Retrieve all chat prompts and check that they match the expected values
	prompts, err := service.GetAllChatPrompts(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve chat prompts: %v", err)
	}
	if len(prompts) != 2 {
		t.Errorf("expected 2 chat prompts, but got %d", len(prompts))
	}
	if prompts[0].ChatSessionUuid != prompt1.ChatSessionUuid || prompts[0].Role != prompt1.Role || prompts[0].Content != prompt1.Content || prompts[0].Score != prompt1.Score || prompts[0].UserID != prompt1.UserID ||
		prompts[1].ChatSessionUuid != prompt2.ChatSessionUuid || prompts[1].Role != prompt2.Role || prompts[1].Content != prompt2.Content || prompts[1].Score != prompt2.Score || prompts[1].UserID != prompt2.UserID {
		t.Error("retrieved chat prompts do not match expected values")
	}

	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatPrompt(context.Background(), prompt1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	if err := service.DeleteChatPrompt(context.Background(), prompt2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	promptsAfterDelete, _ := service.GetAllChatPrompts(context.Background())
	if len(promptsAfterDelete) != 0 {
		t.Error("retrieved chat prompts")
	}
	fmt.Printf("%+v", promptsAfterDelete)

}

func TestGetChatPromptsByTopic(t *testing.T) {

	// Create a new ChatPromptService with the test database connection
	q := sqlc_queries.New(db)
	service := NewChatPromptService(q)

	// Insert two chat prompts into the database with different topics
	prompt1_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic 1", Role: "Test Role 1", Content: "Test Content 1", UserID: 1}
	prompt1, err := service.CreateChatPrompt(context.Background(), prompt1_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}
	prompt2_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic 2", Role: "Test Role 2", Content: "Test Content 2", UserID: 2}
	prompt2, err := service.CreateChatPrompt(context.Background(), prompt2_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	// Retrieve chat prompts by topic and check that they match the expected values
	topic := "Test Topic 1"
	prompts, err := service.GetChatPromptsBySessionUUID(context.Background(), topic)
	if err != nil {
		t.Fatalf("failed to retrieve chat prompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 chat prompt, but got %d", len(prompts))
	}
	if prompts[0].ChatSessionUuid != prompt1.ChatSessionUuid || prompts[0].Role != prompt1.Role || prompts[0].Content != prompt1.Content || prompts[0].Score != prompt1.Score || prompts[0].UserID != prompt1.UserID {
		t.Error("retrieved chat prompts do not match expected values")
	}

	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatPrompt(context.Background(), prompt1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	if err := service.DeleteChatPrompt(context.Background(), prompt2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	promptsAfterDelete, _ := service.GetAllChatPrompts(context.Background())
	if len(promptsAfterDelete) != 0 {
		t.Error("retrieved chat prompts")
	}
	fmt.Printf("%+v", promptsAfterDelete)

}

func TestGetChatPromptsByUserID(t *testing.T) {
	// Create a new ChatPromptService with the test database connection
	q := sqlc_queries.New(db)
	service := NewChatPromptService(q)

	// Insert two chat prompts into the database with different user IDs
	prompt1_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic 1", Role: "Test Role 1", Content: "Test Content 1", UserID: 1}
	prompt1, err := service.CreateChatPrompt(context.Background(), prompt1_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}
	prompt2_params := sqlc_queries.CreateChatPromptParams{ChatSessionUuid: "Test Topic 2", Role: "Test Role 2", Content: "Test Content 2", UserID: 2}
	prompt2, err := service.CreateChatPrompt(context.Background(), prompt2_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	// Retrieve chat prompts by user ID and check that they match the expected values
	userID := int32(1)
	prompts, err := service.GetChatPromptsByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to retrieve chat prompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 chat prompt, but got %d", len(prompts))
	}
	if prompts[0].ChatSessionUuid != prompt1.ChatSessionUuid || prompts[0].Role != prompt1.Role || prompts[0].Content != prompt1.Content || prompts[0].Score != prompt1.Score || prompts[0].UserID != prompt1.UserID {
		t.Error("retrieved chat prompts do not match expected values")
	}

	// Delete the chat prompt and check that it was deleted from the database
	if err := service.DeleteChatPrompt(context.Background(), prompt1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	if err := service.DeleteChatPrompt(context.Background(), prompt2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	promptsAfterDelete, _ := service.GetAllChatPrompts(context.Background())
	if len(promptsAfterDelete) != 0 {
		t.Error("retrieved chat prompts")
	}
	fmt.Printf("%+v", promptsAfterDelete)
}
