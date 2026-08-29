package svc

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestActiveSessionFromRecordKeepsWorkspaceOptional(t *testing.T) {
	now := time.Now()
	withWorkspace := activeSessionFromRecord(sqlc_queries.UserActiveChatSession{
		ID: 1, UserID: 2, ChatSessionUuid: "session", WorkspaceID: sql.NullInt32{Int32: 3, Valid: true}, CreatedAt: now, UpdatedAt: now,
	})
	if withWorkspace.WorkspaceID == nil || *withWorkspace.WorkspaceID != 3 {
		t.Fatalf("workspace ID = %#v, want 3", withWorkspace.WorkspaceID)
	}
	if withWorkspace.SessionUUID != "session" {
		t.Fatalf("session UUID = %q", withWorkspace.SessionUUID)
	}

	withoutWorkspace := activeSessionFromRecord(sqlc_queries.UserActiveChatSession{})
	if withoutWorkspace.WorkspaceID != nil {
		t.Fatalf("workspace ID = %#v, want nil", withoutWorkspace.WorkspaceID)
	}
}

func createActiveSessionTestChatSession(t *testing.T, q *sqlc_queries.Queries, userID, workspaceID int32, uuid string) {
	t.Helper()
	_, err := q.CreateChatSessionInWorkspace(context.Background(), sqlc_queries.CreateChatSessionInWorkspaceParams{
		UserID: userID, Uuid: uuid, Topic: uuid, CreatedAt: time.Now(), Active: true,
		MaxLength: 10, Model: "test-model", WorkspaceID: sql.NullInt32{Int32: workspaceID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetWorkspaceActiveSession(t *testing.T) {
	q, userID := createWorkspaceUser(t, "active-session-success")
	workspace := createWorkspaceRecord(t, q, userID, "active-session-workspace", false)
	createActiveSessionTestChatSession(t, q, userID, workspace.ID, "active-session-chat")

	active, err := NewUserActiveChatSessionService(q).SetWorkspaceActiveSession(context.Background(), SetWorkspaceActiveSessionCommand{
		UserID: userID, WorkspaceUUID: workspace.Uuid, SessionUUID: "active-session-chat",
	})
	if err != nil {
		t.Fatal(err)
	}
	if active.WorkspaceID == nil || *active.WorkspaceID != workspace.ID || active.SessionUUID != "active-session-chat" {
		t.Fatalf("unexpected active session: %#v", active)
	}
}

func TestSetWorkspaceActiveSessionRejectsDifferentWorkspaceOwner(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "active-session-workspace-owner")
	_, otherID := createWorkspaceUser(t, "active-session-workspace-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "foreign-active-workspace", false)
	createActiveSessionTestChatSession(t, q, ownerID, workspace.ID, "foreign-workspace-session")

	_, err := NewUserActiveChatSessionService(q).SetWorkspaceActiveSession(context.Background(), SetWorkspaceActiveSessionCommand{
		UserID: otherID, WorkspaceUUID: workspace.Uuid, SessionUUID: "foreign-workspace-session",
	})
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestSetWorkspaceActiveSessionRejectsDifferentSessionOwner(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "active-session-owner")
	_, otherID := createWorkspaceUser(t, "active-session-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "active-owner-workspace", false)
	createActiveSessionTestChatSession(t, q, otherID, workspace.ID, "foreign-owner-session")

	_, err := NewUserActiveChatSessionService(q).SetWorkspaceActiveSession(context.Background(), SetWorkspaceActiveSessionCommand{
		UserID: ownerID, WorkspaceUUID: workspace.Uuid, SessionUUID: "foreign-owner-session",
	})
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestSetWorkspaceActiveSessionRejectsSessionFromAnotherWorkspace(t *testing.T) {
	q, userID := createWorkspaceUser(t, "active-session-membership")
	requested := createWorkspaceRecord(t, q, userID, "requested-active-workspace", false)
	actual := createWorkspaceRecord(t, q, userID, "actual-active-workspace", false)
	createActiveSessionTestChatSession(t, q, userID, actual.ID, "wrong-workspace-session")

	_, err := NewUserActiveChatSessionService(q).SetWorkspaceActiveSession(context.Background(), SetWorkspaceActiveSessionCommand{
		UserID: userID, WorkspaceUUID: requested.Uuid, SessionUUID: "wrong-workspace-session",
	})
	if !domain.IsKind(err, domain.KindInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}

func TestSetGlobalActiveSessionRejectsDifferentSessionOwner(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "global-active-owner")
	_, otherID := createWorkspaceUser(t, "global-active-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "global-active-workspace", false)
	createActiveSessionTestChatSession(t, q, ownerID, workspace.ID, "global-foreign-session")

	_, err := NewUserActiveChatSessionService(q).SetGlobalActiveSession(context.Background(), SetGlobalActiveSessionCommand{
		UserID: otherID, SessionUUID: "global-foreign-session",
	})
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestGetWorkspaceActiveSessionChecksWorkspaceOwner(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "get-active-owner")
	_, otherID := createWorkspaceUser(t, "get-active-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "get-active-workspace", false)
	createActiveSessionTestChatSession(t, q, ownerID, workspace.ID, "get-active-session")
	service := NewUserActiveChatSessionService(q)
	if _, err := service.SetWorkspaceActiveSession(context.Background(), SetWorkspaceActiveSessionCommand{
		UserID: ownerID, WorkspaceUUID: workspace.Uuid, SessionUUID: "get-active-session",
	}); err != nil {
		t.Fatal(err)
	}

	active, err := service.GetWorkspaceActiveSession(context.Background(), GetWorkspaceActiveSessionQuery{
		UserID: ownerID, WorkspaceUUID: workspace.Uuid,
	})
	if err != nil || active.SessionUUID != "get-active-session" {
		t.Fatalf("active session = %#v, error = %v", active, err)
	}
	_, err = service.GetWorkspaceActiveSession(context.Background(), GetWorkspaceActiveSessionQuery{
		UserID: otherID, WorkspaceUUID: workspace.Uuid,
	})
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

var (
	_ WorkspaceTransactionManager    = (*sqlcTransactionManager)(nil)
	_ SnapshotCopyTransactionManager = (*sqlcTransactionManager)(nil)
)
