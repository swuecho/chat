package svc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestChatPromptService(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatPromptService(q)

	prompt_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic", Role: "Test Role", Content: "Test Content", UserID: 1,
	}
	prompt, err := service.CreateChatPrompt(context.Background(), prompt_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	retrievedPrompt, err := service.GetChatPromptByID(context.Background(), prompt.ID)
	if err != nil {
		t.Fatalf("failed to get chat prompt: %v", err)
	}
	if retrievedPrompt.ChatSessionUuid != prompt.ChatSessionUuid || retrievedPrompt.Role != prompt.Role ||
		retrievedPrompt.Content != prompt.Content || retrievedPrompt.Score != prompt.Score ||
		retrievedPrompt.UserID != prompt.UserID {
		t.Error("retrieved chat prompt does not match expected values")
	}

	updated_params := sqlc_queries.UpdateChatPromptParams{
		ID: prompt.ID, ChatSessionUuid: "Updated Test Topic",
		Role: "Updated Test Role", Content: "Updated Test Content", Score: 0.75,
	}
	if _, err := service.UpdateChatPrompt(context.Background(), updated_params); err != nil {
		t.Fatalf("failed to update chat prompt: %v", err)
	}
	retrievedPrompt, err = service.GetChatPromptByID(context.Background(), prompt.ID)
	if err != nil {
		t.Fatalf("failed to get chat prompt: %v", err)
	}
	if retrievedPrompt.ChatSessionUuid != updated_params.ChatSessionUuid ||
		retrievedPrompt.Role != updated_params.Role ||
		retrievedPrompt.Content != updated_params.Content ||
		retrievedPrompt.Score != updated_params.Score {
		t.Error("retrieved chat prompt does not match expected values")
	}

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
}

func TestGetAllChatPrompts(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatPromptService(q)

	prompt1_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic 1", Role: "Test Role 1", Content: "Test Content 1", UserID: 1,
	}
	prompt1, err := service.CreateChatPrompt(context.Background(), prompt1_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}
	prompt2_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic 2", Role: "Test Role 2", Content: "Test Content 2", UserID: 2,
	}
	prompt2, err := service.CreateChatPrompt(context.Background(), prompt2_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	prompts, err := service.GetAllChatPrompts(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve chat prompts: %v", err)
	}
	if len(prompts) != 2 {
		t.Errorf("expected 2 chat prompts, but got %d", len(prompts))
	}

	if err := service.DeleteChatPrompt(context.Background(), prompt1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	if err := service.DeleteChatPrompt(context.Background(), prompt2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
}

func TestGetChatPromptsByTopic(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatPromptService(q)

	prompt1_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic 1", Role: "Test Role 1", Content: "Test Content 1", UserID: 1,
	}
	prompt1, err := service.CreateChatPrompt(context.Background(), prompt1_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}
	prompt2_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic 2", Role: "Test Role 2", Content: "Test Content 2", UserID: 2,
	}
	prompt2, err := service.CreateChatPrompt(context.Background(), prompt2_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	prompts, err := service.GetChatPromptsBySessionUUID(context.Background(), "Test Topic 1")
	if err != nil {
		t.Fatalf("failed to retrieve chat prompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 chat prompt, but got %d", len(prompts))
	}
	if prompts[0].ChatSessionUuid != prompt1.ChatSessionUuid {
		t.Error("retrieved chat prompts do not match expected values")
	}

	if err := service.DeleteChatPrompt(context.Background(), prompt1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	if err := service.DeleteChatPrompt(context.Background(), prompt2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
}

func TestGetChatPromptsByUserID(t *testing.T) {
	q := sqlc_queries.New(testDB)
	service := NewChatPromptService(q)

	prompt1_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic 1", Role: "Test Role 1", Content: "Test Content 1", UserID: 1,
	}
	prompt1, err := service.CreateChatPrompt(context.Background(), prompt1_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}
	prompt2_params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: "Test Topic 2", Role: "Test Role 2", Content: "Test Content 2", UserID: 2,
	}
	prompt2, err := service.CreateChatPrompt(context.Background(), prompt2_params)
	if err != nil {
		t.Fatalf("failed to create chat prompt: %v", err)
	}

	prompts, err := service.GetChatPromptsByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to retrieve chat prompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 chat prompt, but got %d", len(prompts))
	}
	if prompts[0].ChatSessionUuid != prompt1.ChatSessionUuid {
		t.Error("retrieved chat prompts do not match expected values")
	}

	if err := service.DeleteChatPrompt(context.Background(), prompt1.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}
	if err := service.DeleteChatPrompt(context.Background(), prompt2.ID); err != nil {
		t.Fatalf("failed to delete chat prompt: %v", err)
	}

	promptsAfterDelete, _ := service.GetAllChatPrompts(context.Background())
	if len(promptsAfterDelete) != 0 {
		t.Error("expected all prompts to be deleted")
	}
	fmt.Printf("%+v", promptsAfterDelete)
}
