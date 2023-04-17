package main

import (
	"bytes"
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
func TestListresults(t *testing.T) {
	q := sqlc_queries.New(db)
	h := NewChatAPIHandler(q) // create a new ChatAPIHandler instance for testing
	router := mux.NewRouter()
	h.Register(router)

	// Now let's create our expected results. Create two results and insert them into the database using the queries.
	expectedResults := []sqlc_queries.ChatApi{
		{
			Name:          "Test API 1",
			Label:         "Test Label 1",
			IsDefault:     false,
			Url:           "http://test.url.com",
			ApiAuthHeader: "Authorization",
			ApiAuthKey:    "TestKey1",
		},
		{
			Name:          "Test API 2",
			Label:         "Test Label 2",
			IsDefault:     false,
			Url:           "http://test.url2.com",
			ApiAuthHeader: "Authorization",
			ApiAuthKey:    "TestKey2",
		},
	}

	for _, api := range expectedResults {
		_, err := q.CreateChatAPI(context.Background(), sqlc_queries.CreateChatAPIParams{
			Name:          api.Name,
			Label:         api.Label,
			IsDefault:     api.IsDefault,
			Url:           api.Url,
			ApiAuthHeader: api.ApiAuthHeader,
			ApiAuthKey:    api.ApiAuthKey,
		})
		if err != nil {
			t.Errorf("Error creating test data: %s", err.Error())
		}
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
	var results []sqlc_queries.ChatApi
	body_bytes := rr.Body.Bytes()
	println(body_bytes)
	err = json.Unmarshal(body_bytes, &results)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}

	if len(results) != 2 {
		t.Errorf("expected 2 chat APIs, got %d", len(results))
	}

	// ensure the returned values are what we expect them to be
	for i, api := range expectedResults {
		assert.Equal(t, api.Name, results[i].Name)
		assert.Equal(t, api.Label, results[i].Label)
		assert.Equal(t, api.IsDefault, results[i].IsDefault)
		assert.Equal(t, api.Url, results[i].Url)
		assert.Equal(t, api.ApiAuthHeader, results[i].ApiAuthHeader)
		assert.Equal(t, api.ApiAuthKey, results[i].ApiAuthKey)
	}

	// Now lets update the the first element of our expected results array and call PUT on the endpoint

	expectedResults[0].Name = "Test API 1 Updated"
	expectedResults[0].Label = "Test Label 1 Updated"

	updateBytes, err := json.Marshal(expectedResults[0])
	if err != nil {
		t.Errorf("Error marshaling update payload: %s", err.Error())
	}

	// Create an HTTP request so we can simulate a PUT with the payload
	updateReq, err := http.NewRequest("PUT", fmt.Sprintf("/chat_apis/%d", results[0].ID), bytes.NewBuffer(updateBytes))
	if err != nil {
		t.Fatal(err)
	}

	updateRR := httptest.NewRecorder()

	router.ServeHTTP(updateRR, updateReq)

	if status := updateRR.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// ensure the new values are returned and were also updated in the database
	var updatedResult sqlc_queries.ChatApi
	err = json.Unmarshal(updateRR.Body.Bytes(), &updatedResult)
	if err != nil {
		t.Errorf("Error parsing response body: %s", err.Error())
	}

	assert.Equal(t, expectedResults[0].Name, updatedResult.Name)
	assert.Equal(t, expectedResults[0].Label, updatedResult.Label)
	// And now call the DELETE endpoint to remove all the created ChatAPIs
	deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("/chat_apis/%d", results[0].ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	deleteRR := httptest.NewRecorder()

	router.ServeHTTP(deleteRR, deleteReq)

	if status := deleteRR.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
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
	err = json.Unmarshal(body_bytes, &results)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}
	// len 1
	if len(results) != 1 {
		t.Errorf("expected 1 chat API, got %d", len(results))
	}

	// first results's name is  "Test API 2"
	assert.Equal(t, results[0].Name, "Test API 2")
	// delete all results
	deleteReq2, err := http.NewRequest("DELETE", fmt.Sprintf("/chat_apis/%d", results[0].ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	deleteRR2 := httptest.NewRecorder()

	router.ServeHTTP(deleteRR2, deleteReq2)

	if status := deleteRR2.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// there should be no results
	req, err = http.NewRequest("GET", "/chat_apis", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// ensure that we get an array of one chat API in the response body
	body_bytes = rr.Body.Bytes()
	println(body_bytes)
	err = json.Unmarshal(body_bytes, &results)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}
	// len 0
	if len(results) != 0 {
		t.Errorf("expected 0 chat API, got %d", len(results))
	}
}
