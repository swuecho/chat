package svc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestCreateWorkspaceSessionCommitsCompleteAggregate(t *testing.T) {
	q, userID := createWorkspaceUser(t, "create-session-success")
	workspace := createWorkspaceRecord(t, q, userID, "workspace-session-success", false)
	service := NewChatWorkspaceService(q)
	result, err := service.CreateWorkspaceSession(context.Background(), CreateWorkspaceSessionCommand{
		UserID: userID, WorkspaceUUID: workspace.Uuid, Topic: "Transactional session",
		Model: "test-model", DefaultSystemPrompt: "Stay focused",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Session.WorkspaceID == nil || *result.Session.WorkspaceID != workspace.ID {
		t.Fatalf("session workspace mismatch: %#v", result.Session.WorkspaceID)
	}
	prompt, err := q.GetOneChatPromptBySessionUUID(context.Background(), result.Session.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.Content != "Stay focused" {
		t.Fatalf("prompt = %q", prompt.Content)
	}
	active, err := q.GetUserActiveSession(context.Background(), sqlc_queries.GetUserActiveSessionParams{UserID: userID, Column2: workspace.ID})
	if err != nil {
		t.Fatal(err)
	}
	if active.ChatSessionUuid != result.Session.UUID {
		t.Fatalf("active session = %q", active.ChatSessionUuid)
	}
}

func TestCreateWorkspaceSessionRollsBackOnPromptFailure(t *testing.T) {
	q, userID := createWorkspaceUser(t, "create-session-rollback")
	workspace := createWorkspaceRecord(t, q, userID, "workspace-session-rollback", false)
	service := NewChatWorkspaceService(q)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	service.newID = func() string {
		calls++
		if calls == 2 {
			cancel()
		}
		if calls == 1 {
			return "workspace-session-rolled-back"
		}
		return "workspace-prompt-rolled-back"
	}
	_, err := service.CreateWorkspaceSession(ctx, CreateWorkspaceSessionCommand{
		UserID: userID, WorkspaceUUID: workspace.Uuid, Topic: "Rollback", Model: "test-model",
	})
	if err == nil {
		t.Fatal("expected transaction failure")
	}
	_, lookupErr := q.GetChatSessionByUUIDWithInActive(context.Background(), "workspace-session-rolled-back")
	if !errors.Is(lookupErr, sql.ErrNoRows) {
		t.Fatalf("expected rollback, got %v", lookupErr)
	}
}
