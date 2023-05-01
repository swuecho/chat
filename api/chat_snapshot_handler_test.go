package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"gotest.tools/v3/assert"
)

// the code below do db update directly in instead of using handler, please change to use handler
func TestChatSnapshot(t *testing.T) {
	q := sqlc_queries.New(db)
	h := NewChatSnapshotHandler(q) // create a new ChatSnapshotHandler instance for testing
	router := mux.NewRouter()
	h.Register(router)
	// add a system user
	snapshot_uuid := uuid.NewString()
	userID := 1
	one, err := h.service.q.CreateChatSnapshot(context.Background(), sqlc_queries.CreateChatSnapshotParams{
		Uuid:         snapshot_uuid,
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
	assert.Equal(t, one.Uuid, snapshot_uuid)
	req, err := http.NewRequest("GET", fmt.Sprintf("/uuid/chat_snapshot/%s", one.Uuid), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	body_bytes := rr.Body.Bytes()
	println(body_bytes)
	// test delete snapshot should fail without context,
	// test delete ok with context
	// ctx3 := context.WithValue(deleteReq2.Context(), userContextKey, strconv.Itoa(int(admin.ID)))

}
