package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func createSnapshotFixture(t *testing.T, suffix string) (*sqlc_queries.Queries, int32, string) {
	t.Helper()
	ctx := context.Background()
	q := sqlc_queries.New(testDB)
	user, err := q.CreateAuthUser(ctx, sqlc_queries.CreateAuthUserParams{
		Email: "snapshot-" + suffix + "@test.com", Username: "snapshot-" + suffix,
		Password: "unused",
	})
	if err != nil {
		t.Fatal(err)
	}
	sourceUUID := "snapshot-source-" + suffix
	_, err = q.CreateOrUpdateChatSessionByUUID(ctx, sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: sourceUUID, UserID: user.ID, Topic: "Source", Model: "test-model",
		MaxLength: 12, Temperature: .4, TopP: .9, MaxTokens: 1024, N: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	promptUUID := "snapshot-prompt-" + suffix
	if _, err := q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams{
		Uuid: promptUUID, ChatSessionUuid: sourceUUID, Role: "system", Content: "Be helpful",
		UserID: user.ID, CreatedBy: user.ID, UpdatedBy: user.ID,
	}); err != nil {
		t.Fatal(err)
	}
	conversation, _ := json.Marshal([]snapshotConversationMessage{
		{UUID: promptUUID, Text: "Be helpful"},
		{UUID: "question", Text: "Hello", Inversion: true},
		{UUID: "answer", Text: "Hi there"},
	})
	snapshotUUID := "snapshot-" + suffix
	if _, err := q.CreateChatSnapshot(ctx, sqlc_queries.CreateChatSnapshotParams{
		Uuid: snapshotUUID, UserID: user.ID, Title: "Copied conversation", Model: "test-model",
		Tags: json.RawMessage(`[]`), Conversation: conversation, Session: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatal(err)
	}
	return q, user.ID, snapshotUUID
}

func TestCreateSessionFromSnapshot(t *testing.T) {
	q, userID, snapshotUUID := createSnapshotFixture(t, "success")
	service := NewChatSessionService(q)
	result, err := service.CreateSessionFromSnapshot(context.Background(), CreateSessionFromSnapshotCommand{
		SnapshotUUID: snapshotUUID, UserID: userID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionUUID == "" {
		t.Fatal("expected a session UUID")
	}
	messages, err := q.GetChatMessagesBySessionUUID(context.Background(), sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid: result.SessionUUID, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[0].Role != "user" || messages[1].Role != "assistant" {
		t.Fatalf("unexpected copied messages: %#v", messages)
	}
}

func TestCreateSessionFromSnapshotRollsBack(t *testing.T) {
	q, userID, snapshotUUID := createSnapshotFixture(t, "rollback")
	service := NewChatSessionService(q)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	service.newID = func() string {
		calls++
		if calls == 3 {
			cancel()
		}
		if calls == 1 {
			return "rolled-back-session"
		}
		return "rolled-back-object"
	}

	_, err := service.CreateSessionFromSnapshot(ctx, CreateSessionFromSnapshotCommand{
		SnapshotUUID: snapshotUUID, UserID: userID,
	})
	if err == nil {
		t.Fatal("expected transaction failure")
	}
	_, lookupErr := q.GetChatSessionByUUIDWithInActive(context.Background(), "rolled-back-session")
	if !errors.Is(lookupErr, sql.ErrNoRows) {
		t.Fatalf("expected session rollback, got %v", lookupErr)
	}
}
