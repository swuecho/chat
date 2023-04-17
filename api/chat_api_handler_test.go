package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"gotest.tools/v3/assert"
)

func TestListChatAPIs(t *testing.T) {
	q := sqlc_queries.New(db)
	h := NewChatAPIHandler(q) // create a new ChatAPIHandler instance for testing
	router := mux.NewRouter()
	h.Register(router)

	// create sample chat API data to be retrieved by the handler
	_, err := q.CreateChatAPI(context.Background(), sqlc_queries.CreateChatAPIParams{
		Name:          "Test API 1",
		Label:         "Test Label 1",
		IsDefault:     false,
		Url:           "http://test.url.com",
		ApiAuthHeader: "Authorization",
		ApiAuthKey:    "TestKey1",
	})
	if err != nil {
		t.Errorf("error creating test data: %s", err.Error())
	}

	_, err = q.CreateChatAPI(context.Background(), sqlc_queries.CreateChatAPIParams{
		Name:          "Test API 2",
		Label:         "Test Label 2",
		IsDefault:     false,
		Url:           "http://test.url2.com",
		ApiAuthHeader: "Authorization",
		ApiAuthKey:    "TestKey2",
	})
	if err != nil {
		t.Errorf("error creating test data: %s", err.Error())
	}

	req, err := http.NewRequest("GET", "/chat_apis", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// ensure that we get an array of two chat APIs in the response body
	var chatAPIs []sqlc_queries.ChatApi
	body_bytes := rr.Body.Bytes()
	println(body_bytes)
	err = json.Unmarshal(body_bytes, &chatAPIs)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}

	if len(chatAPIs) != 2 {
		t.Errorf("expected 2 chat APIs, got %d", len(chatAPIs))
	}
	// delete chat
	err = q.DeleteChatAPI(context.Background(), chatAPIs[0].ID)
	if err != nil {
		t.Errorf("error deleting test data: %s", err.Error())
	}

	// only one left
	req, err = http.NewRequest("GET", "/chat_apis", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// ensure that we get an array of one chat API in the response body
	body_bytes = rr.Body.Bytes()
	println(body_bytes)
	err = json.Unmarshal(body_bytes, &chatAPIs)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}
	// len 1
	if len(chatAPIs) != 1 {
		t.Errorf("expected 1 chat API, got %d", len(chatAPIs))
	}

	// first chatAPIs's name is  "Test API 2"
	assert.Equal(t, chatAPIs[0].Name, "Test API 2")
	// update chatAPIs's name to "Test API 2-1"
	chatAPIs[0].Name = "Test API 2-1"
	// update chatAPIs's name to "Test API 2-1"
	chatAPIs[0].Label = "Test Label 2-1"
	// update chatapi to db
	_, err = q.UpdateChatAPI(context.Background(), sqlc_queries.UpdateChatAPIParams{
		ID:            chatAPIs[0].ID,
		Name:          chatAPIs[0].Name,
		Label:         chatAPIs[0].Label,
		IsDefault:     chatAPIs[0].IsDefault,
		Url:           chatAPIs[0].Url,
		ApiAuthHeader: chatAPIs[0].ApiAuthHeader,
		ApiAuthKey:    chatAPIs[0].ApiAuthKey,
	})

	// get chatAPIs's name to "Test API 2-1"
	req, err = http.NewRequest("GET", "/chat_apis", nil)
	if err != nil {
		t.Fatal(err)
	}
	// test updated chatAPIs's name to "Test API 2-1"
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// ensure that we get an array of one chat API in the response body
	body_bytes = rr.Body.Bytes()
	println(body_bytes)
	err = json.Unmarshal(body_bytes, &chatAPIs)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}
	// len 1
	if len(chatAPIs) != 1 {
		t.Errorf("expected 1 chat API, got %d", len(chatAPIs))
	}
	// first chatAPIs's name is  "Test API 2-1"
	assert.Equal(t, chatAPIs[0].Name, "Test API 2-1")
	// delete all chatAPIs
	err = q.DeleteChatAPI(context.Background(), chatAPIs[0].ID)
	if err != nil {
		t.Errorf("error deleting test data: %s", err.Error())
	}
	// there should be no chatAPIs
	req, err = http.NewRequest("GET", "/chat_apis", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// ensure that we get an array of one chat API in the response body
	body_bytes = rr.Body.Bytes()
	println(body_bytes)
	err = json.Unmarshal(body_bytes, &chatAPIs)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}
	// len 0
	if len(chatAPIs) != 0 {
		t.Errorf("expected 0 chat API, got %d", len(chatAPIs))
	}
}
