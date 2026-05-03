package handler

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"

	"github.com/swuecho/chat_backend/middleware"
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
	return context.WithValue(context.Background(), middleware.UserContextKey, strconv.Itoa(userID))
}
