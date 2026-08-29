package svc

import (
	"context"
	"sync"
	"testing"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func createWorkspaceUser(t *testing.T, suffix string) (*sqlc_queries.Queries, int32) {
	t.Helper()
	q := sqlc_queries.New(testDB)
	user, err := q.CreateAuthUser(context.Background(), sqlc_queries.CreateAuthUserParams{
		Email: "workspace-default-" + suffix + "@test.com", Username: "workspace-default-" + suffix,
		Password: "unused",
	})
	if err != nil {
		t.Fatal(err)
	}
	return q, user.ID
}

func createWorkspaceRecord(t *testing.T, q *sqlc_queries.Queries, userID int32, uuid string, isDefault bool) sqlc_queries.ChatWorkspace {
	t.Helper()
	workspace, err := q.CreateWorkspace(context.Background(), sqlc_queries.CreateWorkspaceParams{
		Uuid: uuid, UserID: userID, Name: uuid, Color: "#6366f1", Icon: "folder", IsDefault: isDefault,
	})
	if err != nil {
		t.Fatal(err)
	}
	return workspace
}

func TestSetWorkspaceAsDefaultForUser(t *testing.T) {
	q, userID := createWorkspaceUser(t, "success")
	createWorkspaceRecord(t, q, userID, "default-original", true)
	target := createWorkspaceRecord(t, q, userID, "default-target", false)

	workspace, err := NewChatWorkspaceService(q).SetWorkspaceAsDefaultForUser(context.Background(), SetDefaultWorkspaceCommand{
		UserID: userID, WorkspaceUUID: target.Uuid,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !workspace.IsDefault {
		t.Fatal("expected target workspace to be default")
	}
	all, err := q.GetWorkspacesByUserID(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	defaults := 0
	for _, item := range all {
		if item.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		t.Fatalf("expected exactly one default, got %d", defaults)
	}
}

func TestSetWorkspaceAsDefaultRejectsDifferentOwner(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "owner")
	_, otherID := createWorkspaceUser(t, "other")
	target := createWorkspaceRecord(t, q, ownerID, "owner-workspace", true)

	_, err := NewChatWorkspaceService(q).SetWorkspaceAsDefaultForUser(context.Background(), SetDefaultWorkspaceCommand{
		UserID: otherID, WorkspaceUUID: target.Uuid,
	})
	if !domain.IsKind(err, domain.KindForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestConcurrentDefaultWorkspaceChangesKeepSingleDefault(t *testing.T) {
	q, userID := createWorkspaceUser(t, "concurrent")
	createWorkspaceRecord(t, q, userID, "concurrent-original", true)
	first := createWorkspaceRecord(t, q, userID, "concurrent-first", false)
	second := createWorkspaceRecord(t, q, userID, "concurrent-second", false)
	service := NewChatWorkspaceService(q)

	var wait sync.WaitGroup
	for _, workspace := range []sqlc_queries.ChatWorkspace{first, second} {
		wait.Add(1)
		go func(uuid string) {
			defer wait.Done()
			_, _ = service.SetWorkspaceAsDefaultForUser(context.Background(), SetDefaultWorkspaceCommand{
				UserID: userID, WorkspaceUUID: uuid,
			})
		}(workspace.Uuid)
	}
	wait.Wait()

	all, err := q.GetWorkspacesByUserID(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	defaults := 0
	for _, item := range all {
		if item.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		t.Fatalf("expected exactly one default after concurrent changes, got %d", defaults)
	}
}
