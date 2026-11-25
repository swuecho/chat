package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func createClearTestMessage(t *testing.T, service *ChatMessageService, sessionUUID, uuid string, userID int32) ChatMessage {
	t.Helper()
	message, err := service.CreateChatMessage(context.Background(), CreateChatMessageInput{
		UUID: uuid, ChatSessionUUID: sessionUUID, Role: "user", Content: uuid,
		UserID: userID, CreatedBy: userID, UpdatedBy: userID,
		Raw: json.RawMessage(`{}`), Artifacts: json.RawMessage(`[]`), SuggestedQuestions: json.RawMessage(`[]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	return message
}

func TestClearSessionMessagesSucceedsWhenNothingToClear(t *testing.T) {
	q, ownerID := createWorkspaceUser(t, "clear-owner")
	_, otherID := createWorkspaceUser(t, "clear-other")
	workspace := createWorkspaceRecord(t, q, ownerID, "clear-workspace", false)
	createActiveSessionTestChatSession(t, q, ownerID, workspace.ID, "clear-session")
	service := NewChatMessageService(q)
	ctx := context.Background()

	if err := service.DeleteChatMessagesBySessionUUID(ctx, DeleteSessionMessagesCommand{SessionUUID: "clear-session", UserID: ownerID}); err != nil {
		t.Fatalf("clearing an empty session: %v", err)
	}

	createClearTestMessage(t, service, "clear-session", "clear-message", ownerID)
	if err := service.DeleteChatMessagesBySessionUUID(ctx, DeleteSessionMessagesCommand{SessionUUID: "clear-session", UserID: otherID}); !domain.IsKind(err, domain.KindNotFound) {
		t.Fatalf("cross-owner clear error = %v, want not found", err)
	}
	if count, err := service.GetChatMessagesCount(ctx, ownerID); err != nil || count != 1 {
		t.Fatalf("message count after cross-owner clear = %d, %v; want 1", count, err)
	}

	if err := service.DeleteChatMessagesBySessionUUID(ctx, DeleteSessionMessagesCommand{SessionUUID: "clear-session", UserID: ownerID}); err != nil {
		t.Fatalf("clearing a session with messages: %v", err)
	}
	if err := service.DeleteChatMessagesBySessionUUID(ctx, DeleteSessionMessagesCommand{SessionUUID: "clear-session", UserID: ownerID}); err != nil {
		t.Fatalf("clearing an already cleared session: %v", err)
	}
}

func TestWorkspaceSessionsListMostRecentActivityFirst(t *testing.T) {
	q, userID := createWorkspaceUser(t, "session-order")
	workspace := createWorkspaceRecord(t, q, userID, "session-order-workspace", false)
	ctx := context.Background()
	messages := NewChatMessageService(q)
	createSession := func(uuid string) {
		t.Helper()
		session, err := q.CreateChatSessionInWorkspace(ctx, sqlc_queries.CreateChatSessionInWorkspaceParams{
			UserID: userID, Uuid: uuid, Topic: uuid, Active: true,
			MaxLength: 10, Model: "test-model", WorkspaceID: sql.NullInt32{Int32: workspace.ID, Valid: true},
		})
		if err != nil {
			t.Fatal(err)
		}
		if session.CreatedAt.Year() < 2000 {
			t.Fatalf("session created_at = %v, want the insert time", session.CreatedAt)
		}
	}

	// An empty session created before the active one must not outrank it.
	createSession("order-empty-old")
	createSession("order-active")
	createClearTestMessage(t, messages, "order-active", "order-active-message", userID)
	deleted := createClearTestMessage(t, messages, "order-empty-old", "order-deleted-message", userID)
	if err := messages.DeleteChatMessage(ctx, DeleteChatMessageCommand{ID: deleted.ID, UserID: userID}); err != nil {
		t.Fatal(err)
	}
	createSession("order-empty-new")

	sessions, err := q.GetSessionsByWorkspaceID(ctx, sql.NullInt32{Int32: workspace.ID, Valid: true})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(sessions))
	for i, s := range sessions {
		got[i] = s.Uuid
	}
	want := []string{"order-empty-new", "order-active", "order-empty-old"}
	if len(got) != len(want) {
		t.Fatalf("sessions = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sessions = %v, want %v", got, want)
		}
	}
}

func TestSnapshotSearchMatchesUnstemmedWords(t *testing.T) {
	q, userID := createWorkspaceUser(t, "snapshot-search")
	ctx := context.Background()
	if _, err := q.CreateChatSnapshot(ctx, sqlc_queries.CreateChatSnapshotParams{
		Uuid: "snapshot-search-calculator", UserID: userID, Title: "Interactive Scientific Calculator",
		Model: "test-model", Tags: json.RawMessage(`{}`), Conversation: json.RawMessage(`[]`),
		Session: json.RawMessage(`{}`), Text: "building a calculator",
	}); err != nil {
		t.Fatal(err)
	}

	for _, term := range []string{"calculator", "Calculator", "scientific"} {
		results, err := q.ChatSnapshotSearch(ctx, sqlc_queries.ChatSnapshotSearchParams{UserID: userID, Search: term})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || results[0].Uuid != "snapshot-search-calculator" {
			t.Fatalf("search %q returned %+v, want the calculator snapshot", term, results)
		}
	}
}
