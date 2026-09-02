package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
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

	if rows, err := q.MarkChatRequestStreaming(ctx, sqlc_queries.MarkChatRequestStreamingParams(claim)); err != nil || rows != 1 {
		t.Fatalf("MarkChatRequestStreaming() error = %v", err)
	}
	service := NewChatService(q, "", "")
	if err := service.MarkChatRequestStreaming(ctx, claim.Uuid, claim.ChatSessionUuid, claim.UserID); !errors.Is(err, ErrInvalidChatRequestTransition) {
		t.Fatalf("second streaming transition error = %v, want ErrInvalidChatRequestTransition", err)
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

func TestChatRequestConcurrentClaimHasSingleWinner(t *testing.T) {
	ctx := context.Background()
	q := sqlc_queries.New(testDB)
	user, err := q.CreateAuthUser(ctx, sqlc_queries.CreateAuthUserParams{Email: "request-race@test.com", Username: "request-race", Password: "test"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteAuthUser(ctx, user.Email)
	session, err := q.CreateChatSession(ctx, sqlc_queries.CreateChatSessionParams{UserID: user.ID, Uuid: "request-race-session", Topic: "race", Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	claim := sqlc_queries.ClaimChatRequestParams{Uuid: "request-race-uuid", ChatSessionUuid: session.Uuid, UserID: user.ID}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := q.ClaimChatRequest(ctx, claim); errs <- err }()
	}
	wg.Wait()
	close(errs)
	winners, rejected := 0, 0
	for err := range errs {
		if err == nil {
			winners++
		} else if errors.Is(err, sql.ErrNoRows) {
			rejected++
		} else {
			t.Fatalf("unexpected claim error: %v", err)
		}
	}
	if winners != 1 || rejected != 1 {
		t.Fatalf("winners=%d rejected=%d", winners, rejected)
	}
}

func TestInvalidCompletionDoesNotInsertMessage(t *testing.T) {
	ctx := context.Background()
	q := sqlc_queries.New(testDB)
	user, err := q.CreateAuthUser(ctx, sqlc_queries.CreateAuthUserParams{Email: "invalid-transition@test.com", Username: "invalid-transition", Password: "test"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteAuthUser(ctx, user.Email)
	session, err := q.CreateChatSession(ctx, sqlc_queries.CreateChatSessionParams{UserID: user.ID, Uuid: "invalid-transition-session", Topic: "invalid", Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = q.CompleteChatRequestWithMessage(ctx, sqlc_queries.CompleteChatRequestWithMessageParams{ChatSessionUuid: session.Uuid, MessageUuid: "orphan-answer", Content: "must not persist", Model: "test-model", UserID: user.ID, CreatedBy: user.ID, UpdatedBy: user.ID, Raw: json.RawMessage(`{}`), Artifacts: json.RawMessage(`[]`), SuggestedQuestions: json.RawMessage(`[]`), RequestUuid: "missing-request"})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("completion error=%v, want sql.ErrNoRows", err)
	}
	_, err = q.GetChatMessageByUUID(ctx, sqlc_queries.GetChatMessageByUUIDParams{Uuid: "orphan-answer", UserID: user.ID})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("orphan message lookup error=%v, want sql.ErrNoRows", err)
	}
}
