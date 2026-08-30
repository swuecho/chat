package svc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func saveSessionTestCommand(userID int32, workspaceUUID, sessionUUID string) SaveSessionCommand {
	return SaveSessionCommand{UserID: userID, WorkspaceUUID: workspaceUUID, SessionUUID: sessionUUID,
		Topic: "Saved session", Model: "test-model", MaxLength: 10, Temperature: 0.7,
		MaxTokens: 1000, TopP: 1, N: 1, EnsureSystemPrompt: true,
		DefaultSystemPrompt: "Be precise", ActivateGlobally: true}
}

func TestSaveSessionCommitsSessionPromptAndActiveSelection(t *testing.T) {
	q, userID := createWorkspaceUser(t, "save-session-success")
	workspace := createWorkspaceRecord(t, q, userID, "save-session-workspace", false)
	service := NewChatSessionService(q)

	session, err := service.SaveSession(context.Background(), saveSessionTestCommand(userID, workspace.Uuid, "saved-session"))
	if err != nil {
		t.Fatal(err)
	}
	if session.WorkspaceID == nil || *session.WorkspaceID != workspace.ID {
		t.Fatalf("workspace ID = %#v, want %d", session.WorkspaceID, workspace.ID)
	}
	prompt, err := q.GetOneChatPromptBySessionUUID(context.Background(), session.UUID)
	if err != nil || prompt.Content != "Be precise" {
		t.Fatalf("prompt = %#v, error = %v", prompt, err)
	}
	active, err := q.GetUserActiveSession(context.Background(), sqlc_queries.GetUserActiveSessionParams{
		UserID: userID, WorkspaceID: sql.NullInt32{},
	})
	if err != nil || active.ChatSessionUuid != session.UUID {
		t.Fatalf("active session = %#v, error = %v", active, err)
	}
}

func TestSaveSessionRejectsWorkspaceOwnedByAnotherUser(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "save-session-workspace-owner")
	_, otherID := createWorkspaceUser(t, "save-session-workspace-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "foreign-save-workspace", false)

	_, err := NewChatSessionService(q).SaveSession(context.Background(), saveSessionTestCommand(otherID, workspace.Uuid, "forbidden-save-session"))
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if _, lookupErr := q.GetChatSessionByUUIDWithInActive(context.Background(), "forbidden-save-session"); !errors.Is(lookupErr, sql.ErrNoRows) {
		t.Fatalf("forbidden save created a session: %v", lookupErr)
	}
}

func TestSaveSessionRejectsExistingSessionOwnedByAnotherUser(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "save-session-owner")
	_, otherID := createWorkspaceUser(t, "save-session-other")
	ownerWorkspace := createWorkspaceRecord(t, q, ownerID, "save-owner-workspace", false)
	otherWorkspace := createWorkspaceRecord(t, q, otherID, "save-other-workspace", false)
	if _, err := NewChatSessionService(q).SaveSession(context.Background(), saveSessionTestCommand(ownerID, ownerWorkspace.Uuid, "owned-save-session")); err != nil {
		t.Fatal(err)
	}

	command := saveSessionTestCommand(otherID, otherWorkspace.Uuid, "owned-save-session")
	command.Topic = "Hijacked"
	_, err := NewChatSessionService(q).SaveSession(context.Background(), command)
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	stored, lookupErr := q.GetChatSessionByUUIDWithInActive(context.Background(), "owned-save-session")
	if lookupErr != nil || stored.Topic == "Hijacked" {
		t.Fatalf("cross-owner save changed session: %#v, error = %v", stored, lookupErr)
	}
}

func TestSaveSessionRollsBackWhenPromptCreationFails(t *testing.T) {
	q, userID := createWorkspaceUser(t, "save-session-rollback")
	workspace := createWorkspaceRecord(t, q, userID, "save-session-rollback-workspace", false)
	service := NewChatSessionService(q)
	ctx, cancel := context.WithCancel(context.Background())
	service.newID = func() string {
		cancel()
		return "cancelled-save-prompt"
	}

	_, err := service.SaveSession(ctx, saveSessionTestCommand(userID, workspace.Uuid, "rolled-back-save-session"))
	if err == nil {
		t.Fatal("expected save failure")
	}
	if _, lookupErr := q.GetChatSessionByUUIDWithInActive(context.Background(), "rolled-back-save-session"); !errors.Is(lookupErr, sql.ErrNoRows) {
		t.Fatalf("expected session rollback, got %v", lookupErr)
	}
}
