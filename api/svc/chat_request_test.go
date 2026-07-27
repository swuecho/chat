package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestChatRequestLifecycleIsIdempotent(t *testing.T) {
	ctx := context.Background()
	q := sqlc_queries.New(testDB)

	user, err := q.CreateAuthUser(ctx, sqlc_queries.CreateAuthUserParams{
		Email: "request-lifecycle@test.com", Username: "request-lifecycle", Password: "test",
	})
	if err != nil {
		t.Fatalf("CreateAuthUser() error = %v", err)
	}
	defer func() {
		if err := q.DeleteAuthUser(ctx, user.Email); err != nil {
			t.Errorf("DeleteAuthUser() cleanup error = %v", err)
		}
	}()
	session, err := q.CreateChatSession(ctx, sqlc_queries.CreateChatSessionParams{
		UserID: user.ID, Uuid: "request-lifecycle-session", Topic: "Request lifecycle", Model: "test-model",
	})
	if err != nil {
		t.Fatalf("CreateChatSession() error = %v", err)
	}
	defer func() {
		if err := q.DeleteChatSession(ctx, session.ID); err != nil {
			t.Errorf("DeleteChatSession() cleanup error = %v", err)
		}
	}()

	claim := sqlc_queries.ClaimChatRequestParams{
		Uuid: "request-lifecycle-uuid", ChatSessionUuid: session.Uuid, UserID: user.ID,
	}
	if _, err := q.ClaimChatRequest(ctx, claim); err != nil {
		t.Fatalf("ClaimChatRequest() error = %v", err)
	}
	if _, err := q.ClaimChatRequest(ctx, claim); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("concurrent claim error = %v, want sql.ErrNoRows", err)
	}

	if err := q.MarkChatRequestStreaming(ctx, sqlc_queries.MarkChatRequestStreamingParams(claim)); err != nil {
		t.Fatalf("MarkChatRequestStreaming() error = %v", err)
	}
	message, err := q.CompleteChatRequestWithMessage(ctx, sqlc_queries.CompleteChatRequestWithMessageParams{
		ChatSessionUuid:    session.Uuid,
		MessageUuid:        "request-lifecycle-answer",
		Content:            "durable answer",
		Model:              "test-model",
		UserID:             user.ID,
		CreatedBy:          user.ID,
		UpdatedBy:          user.ID,
		Raw:                json.RawMessage(`{}`),
		Artifacts:          json.RawMessage(`[]`),
		SuggestedQuestions: json.RawMessage(`[]`),
		RequestUuid:        claim.Uuid,
	})
	if err != nil {
		t.Fatalf("CompleteChatRequestWithMessage() error = %v", err)
	}

	request, err := q.GetChatRequest(ctx, sqlc_queries.GetChatRequestParams(claim))
	if err != nil {
		t.Fatalf("GetChatRequest() error = %v", err)
	}
	if request.Status != "completed" || request.AssistantUuid != message.Uuid {
		t.Fatalf("unexpected completed request: %#v", request)
	}
	if _, err := q.ClaimChatRequest(ctx, claim); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("completed retry claim error = %v, want sql.ErrNoRows", err)
	}
}
