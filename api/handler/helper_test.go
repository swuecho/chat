package handler

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/swuecho/chat_backend/requestctx"
	"github.com/swuecho/chat_backend/testutil"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	db, cleanup := testutil.NewTestDB(m)
	defer cleanup()
	testDB = db
	os.Exit(m.Run())
}

// getContextWithUser creates a background context with the given user ID.
func getContextWithUser(userID int) context.Context {
	return requestctx.WithPrincipal(context.Background(), requestctx.Principal{UserID: int32(userID), Role: "user"})
}
