package svc

import (
	"context"
	"encoding/json"
	"testing"
)

func TestWorkspaceMutationsAreOwnershipScoped(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "workspace-owner")
	_, otherID := createWorkspaceUser(t, "workspace-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "owned-workspace", false)
	service := NewChatWorkspaceService(q)

	if _, err := service.UpdateWorkspaceOrder(context.Background(), UpdateWorkspaceOrderCommand{
		WorkspaceUUID: workspace.Uuid, UserID: otherID, OrderPosition: 99,
	}); err == nil {
		t.Fatal("expected cross-owner reorder to fail")
	}
	if err := service.DeleteWorkspace(context.Background(), DeleteWorkspaceCommand{
		WorkspaceUUID: workspace.Uuid, UserID: otherID,
	}); err == nil {
		t.Fatal("expected cross-owner delete to fail")
	}

	stored, err := q.GetWorkspaceByUUID(context.Background(), workspace.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if stored.OrderPosition == 99 {
		t.Fatal("cross-owner reorder changed workspace")
	}
}

func TestPromptMutationsAreOwnershipScoped(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "prompt-owner")
	_, otherID := createWorkspaceUser(t, "prompt-other")
	service := NewChatPromptService(q)
	prompt, err := service.CreateChatPrompt(context.Background(), CreateChatPromptInput{
		UUID: "owned-prompt", ChatSessionUUID: "owned-prompt-session", Role: "system",
		Content: "original", UserID: ownerID, CreatedBy: ownerID, UpdatedBy: ownerID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateChatPromptByUUID(context.Background(), UpdateChatPromptByUUIDCommand{
		UUID: prompt.UUID, Content: "hijacked", UserID: otherID,
	}); err == nil {
		t.Fatal("expected cross-owner update to fail")
	}
	if err := service.DeleteChatPrompt(context.Background(), DeleteChatPromptCommand{ID: prompt.ID, UserID: otherID}); err == nil {
		t.Fatal("expected cross-owner delete to fail")
	}
	stored, err := service.GetChatPromptByID(context.Background(), prompt.ID, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Content != "original" {
		t.Fatalf("content changed to %q", stored.Content)
	}
}

func TestMessageMutationsAreOwnershipScoped(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "message-owner")
	_, otherID := createWorkspaceUser(t, "message-other")
	service := NewChatMessageService(q)
	message, err := service.CreateChatMessage(context.Background(), CreateChatMessageInput{
		UUID: "owned-message", ChatSessionUUID: "owned-message-session", Role: "user",
		Content: "original", UserID: ownerID, CreatedBy: ownerID, UpdatedBy: ownerID,
		Raw: json.RawMessage(`{}`), Artifacts: json.RawMessage(`[]`), SuggestedQuestions: json.RawMessage(`[]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateChatMessageByUUID(context.Background(), UpdateChatMessageByUUIDInput{
		UUID: message.UUID, Content: "hijacked", UserID: otherID,
	}); err == nil {
		t.Fatal("expected cross-owner update to fail")
	}
	chatService := NewChatService(q, "", "")
	if err := chatService.UpdateChatMessageContent(context.Background(), message.UUID, otherID, "hijacked"); err == nil {
		t.Fatal("expected cross-owner regeneration update to fail")
	}
	if _, err := service.UpdateSuggestedQuestions(context.Background(), message.UUID, otherID, json.RawMessage(`["hijacked"]`)); err == nil {
		t.Fatal("expected cross-owner suggestion update to fail")
	}
	if err := service.DeleteChatMessage(context.Background(), DeleteChatMessageCommand{ID: message.ID, UserID: otherID}); err == nil {
		t.Fatal("expected cross-owner delete to fail")
	}
	stored, err := service.GetChatMessageByID(context.Background(), message.ID, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Content != "original" {
		t.Fatalf("content changed to %q", stored.Content)
	}
}
