package svc

import (
	"database/sql"
	"testing"
	"time"

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

var (
	_ WorkspaceTransactionManager    = (*sqlcTransactionManager)(nil)
	_ SnapshotCopyTransactionManager = (*sqlcTransactionManager)(nil)
)
