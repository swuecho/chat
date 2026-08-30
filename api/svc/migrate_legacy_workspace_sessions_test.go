package svc

import (
	"context"
	"database/sql"
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestMigrateLegacyWorkspaceSessionsRepairsActiveSelection(t *testing.T) {
	q, userID := createWorkspaceUser(t, "legacy-migration-use-case")
	const sessionUUID = "legacy-migration-session"
	if _, err := q.CreateOrUpdateChatSessionByUUID(context.Background(), sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: sessionUUID, UserID: userID, Topic: "Legacy", MaxLength: 10, Model: "test-model", N: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := q.UpsertUserActiveSession(context.Background(), sqlc_queries.UpsertUserActiveSessionParams{
		UserID: userID, ChatSessionUuid: sessionUUID,
	}); err != nil {
		t.Fatal(err)
	}

	result, err := NewChatWorkspaceService(q).MigrateLegacyWorkspaceSessions(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.HasLegacySessions || result.MigratedCount != 1 || result.DefaultWorkspace.ID == 0 {
		t.Fatalf("unexpected migration result: %#v", result)
	}
	remaining, err := q.GetSessionsWithoutWorkspace(context.Background(), userID)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("legacy sessions remaining = %d, error = %v", len(remaining), err)
	}
	active, err := q.GetUserActiveSession(context.Background(), sqlc_queries.GetUserActiveSessionParams{
		UserID: userID, WorkspaceID: sql.NullInt32{Int32: result.DefaultWorkspace.ID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if active.ChatSessionUuid != sessionUUID {
		t.Fatalf("active session = %q, want %q", active.ChatSessionUuid, sessionUUID)
	}
}
