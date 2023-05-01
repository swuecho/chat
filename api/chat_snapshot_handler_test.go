package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"gotest.tools/v3/assert"
)

func getContext(userID int) context.Context {
	return context.WithValue(context.Background(), userContextKey, strconv.Itoa(userID))
    }
// the code below do db update directly in instead of using handler, please change to use handler
func TestChatSnapshot(t *testing.T) {
	const snapshotPath = "/uuid/chat_snapshot/%s"

	q := sqlc_queries.New(db)
	service := NewChatMessageService(q)
	h := NewChatSnapshotHandler(service) // create a new ChatSnapshotHandler instance for testing
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
	req, err := http.NewRequest("GET", fmt.Sprintf(snapshotPath, one.Uuid), nil)
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
	reqDelete, err := http.NewRequest("DELETE", fmt.Sprintf(snapshotPath, one.Uuid), nil)
	if err != nil {
		t.Fatal(err)
	}

	rDelete := httptest.NewRecorder()

	router.ServeHTTP(rDelete, reqDelete)

	if status := rDelete.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	reqDeleteWitUserContext, err := http.NewRequest("DELETE", fmt.Sprintf(snapshotPath, one.Uuid), nil)
	if err != nil {
		t.Fatal(err)
	}

	rDelete3 := httptest.NewRecorder()

	ctx := getContext(userID)
	router.ServeHTTP(rDelete3, reqDeleteWitUserContext.WithContext(ctx))

	if status := rDelete3.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
