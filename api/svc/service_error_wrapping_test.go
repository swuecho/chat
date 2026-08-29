package svc

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/testutil"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	db, cleanup := testutil.NewTestDB(m)
	defer cleanup()
	testDB = db
	os.Exit(m.Run())
}

// --- Service error wrapping tests ---

func TestServiceErrorWrappingNil(t *testing.T) {
	q := sqlc_queries.New(testDB)
	ctx := context.Background()

	testUser, err := NewAuthUserService(q, "test-secret", 100).CreateAuthUser(ctx,
		CreateAuthUserInput{
			Email:    "wrap-test@test.com",
			Username: "wraptest",
			Password: "pbkdf2_sha256$260000$test$test",
		})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	testUID := testUser.ID

	t.Run("ChatSessionService_success_returns_nil_error", func(t *testing.T) {
		svc := NewChatSessionService(q)
		session, err := svc.CreateOrUpdateChatSessionByUUID(ctx,
			CreateOrUpdateChatSessionInput{
				UUID: "test-wrap-nil-session", UserID: testUID,
				Topic: "Wrap Nil Test", Model: "gpt-3.5-turbo", MaxLength: 10,
			})
		if err != nil {
			t.Fatalf("CreateOrUpdateChatSessionByUUID failed: %v", err)
		}
		if session.UUID == "" {
			t.Fatal("expected non-empty session UUID")
		}
		retrieved, err := svc.GetChatSessionByUUID(ctx, session.UUID)
		if err != nil {
			t.Fatalf("GetChatSessionByUUID failed: %v", err)
		}
		if retrieved.UUID != session.UUID {
			t.Fatalf("UUID mismatch: got %s, want %s", retrieved.UUID, session.UUID)
		}
	})

	t.Run("ChatWorkspaceService_success_returns_nil_error", func(t *testing.T) {
		svc := NewChatWorkspaceService(q)
		ws, err := svc.CreateWorkspace(ctx, CreateWorkspaceInput{
			Uuid: "test-wrap-nil-workspace", UserID: testUID,
			Name: "Wrap Nil Test WS", Color: "#6366f1", Icon: "folder",
		})
		if err != nil {
			t.Fatalf("CreateWorkspace failed: %v", err)
		}
		if ws.UUID == "" {
			t.Fatal("expected non-empty workspace UUID")
		}
		retrieved, err := svc.GetWorkspaceByUUID(ctx, ws.UUID)
		if err != nil {
			t.Fatalf("GetWorkspaceByUUID failed: %v", err)
		}
		if retrieved.UUID != ws.UUID {
			t.Fatalf("UUID mismatch: got %s, want %s", retrieved.UUID, ws.UUID)
		}
		if err := svc.DeleteWorkspace(ctx, DeleteWorkspaceCommand{WorkspaceUUID: ws.UUID, UserID: ws.UserID}); err != nil {
			t.Fatalf("DeleteWorkspace failed: %v", err)
		}
	})

	t.Run("AuthUserService_success_returns_nil_error", func(t *testing.T) {
		svc := NewAuthUserService(q, "test-secret", 100)
		stats, total, err := svc.GetUserStats(ctx, PageRequest{Page: 1, Size: 10}, 100)
		if err != nil {
			t.Fatalf("GetUserStats failed: %v", err)
		}
		_ = stats
		_ = total
	})

	t.Run("ChatMessageService_success_returns_nil_error", func(t *testing.T) {
		svc := NewChatMessageService(q)
		msgs, err := svc.GetAllChatMessages(ctx, testUID)
		if err != nil {
			t.Fatalf("GetAllChatMessages failed: %v", err)
		}
		_ = msgs
	})
}

func TestSessionServiceErrorWrappingNonNil(t *testing.T) {
	q := sqlc_queries.New(testDB)
	svc := NewChatSessionService(q)
	ctx := context.Background()

	_, err := svc.GetChatSessionByID(ctx, -1)
	if err == nil {
		t.Fatal("expected error for non-existent session ID")
	}
	if !isErrorWithCode(err, "RES_001") && !isErrorWithCode(err, "DB_001") {
		t.Logf("error type: %T, message: %v", err, err)
	}
}

func isErrorWithCode(err error, code string) bool {
	var apiErr dto.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == code
	}
	return false
}
