package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"gotest.tools/v3/assert"
)

func TestChatSnapshot(t *testing.T) {
	const snapshotPath = "/uuid/chat_snapshot/%s"

	q := sqlc_queries.New(testDB)
	h := NewChatSnapshotHandler(q)

	router := mux.NewRouter()
	h.Register(router)

	userID := 1
	snapshotUUID := NewUUID()

	snapshot, err := h.Service.Q().CreateChatSnapshot(context.Background(), sqlc_queries.CreateChatSnapshotParams{
		Uuid:         snapshotUUID,
		Model:        "gpt3",
		Title:        "test chat snapshot",
		UserID:       int32(userID),
		Session:      json.RawMessage([]byte("{}")),
		Tags:         json.RawMessage([]byte("{}")),
		Text:         "test chat snapshot text",
		Conversation: json.RawMessage([]byte("{}")),
	})
	if err != nil {
		t.Fatalf("failed to create snapshot: %v", err)
	}
	assert.Equal(t, snapshot.Uuid, snapshotUUID)

	// Test GET
	req, _ := http.NewRequest("GET", fmt.Sprintf(snapshotPath, snapshot.Uuid), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test DELETE without auth
	reqDelete, _ := http.NewRequest("DELETE", fmt.Sprintf(snapshotPath, snapshot.Uuid), nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, reqDelete)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Test DELETE with auth
	reqDeleteWithAuth, _ := http.NewRequest("DELETE", fmt.Sprintf(snapshotPath, snapshot.Uuid), nil)
	ctx := getContextWithUser(userID)
	reqDeleteWithAuth = reqDeleteWithAuth.WithContext(ctx)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, reqDeleteWithAuth)
	assert.Equal(t, http.StatusOK, rr.Code)
}
