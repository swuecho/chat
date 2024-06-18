package main

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

// the code below do db update directly in instead of using handler, please change to use handler

// TestChatSnapshot tests the ChatSnapshotHandler
func TestChatSnapshot(t *testing.T) {
	const snapshotPath = "/uuid/chat_snapshot/%s" // API path for snapshots

	// Create a chat service for testing
	q := sqlc_queries.New(db)
	h := NewChatSnapshotHandler(q) // Create a ChatSnapshotHandler

	// Register snapshot API routes
	router := mux.NewRouter()
	h.Register(router)

	// Add a test user
	userID := 1

	// Generate a random UUID for the snapshot
	snapshotUUID := NewUUID()

	// Create a test snapshot
	snapshot, err := h.service.q.CreateChatSnapshot(context.Background(), sqlc_queries.CreateChatSnapshotParams{
		Uuid:         snapshotUUID, // Use the generated UUID
		Model:        "gpt3",
		Title:        "test chat snapshot",
		UserID:       int32(userID),
		Session:      json.RawMessage([]byte("{}")),
		Tags:         json.RawMessage([]byte("{}")),
		Text:         "test chat snapshot text",
		Conversation: json.RawMessage([]byte("{}")),
	})
	if err != nil {
		return
	}
	assert.Equal(t, snapshot.Uuid, snapshotUUID)

	// Test GET snapshot - should succeed
	req, _ := http.NewRequest("GET", fmt.Sprintf(snapshotPath, snapshot.Uuid), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test DELETE snapshot without auth - should fail
	reqDelete, _ := http.NewRequest("DELETE", fmt.Sprintf(snapshotPath, snapshot.Uuid), nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, reqDelete)
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Test DELETE snapshot with auth - should succeed
	reqDeleteWithAuth, _ := http.NewRequest("DELETE", fmt.Sprintf(snapshotPath, snapshot.Uuid), nil)
	ctx := getContextWithUser(userID) // Get auth context
	reqDeleteWithAuth = reqDeleteWithAuth.WithContext(ctx)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, reqDeleteWithAuth)
	assert.Equal(t, http.StatusOK, rr.Code)

}
