package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"gotest.tools/v3/assert"
)

func createTwoChatModel(q *sqlc_queries.Queries) (sqlc_queries.AuthUser, []sqlc_queries.ChatModel) {
	// add a system user
	admin, err := q.CreateAuthUser(context.Background(), sqlc_queries.CreateAuthUserParams{
		Email:       "admin@a.com",
		Username:    "test",
		Password:    "test",
		IsSuperuser: true,
	})

	if err != nil {
		fmt.Printf("Error creating test data: %s", err.Error())
	}
	expectedResults := []sqlc_queries.ChatModel{
		{
			Name:          "Test API 1",
			Label:         "Test Label 1",
			IsDefault:     false,
			Url:           "http://test.url.com",
			ApiAuthHeader: "Authorization",
			ApiAuthKey:    "TestKey1",
			UserID:        admin.ID,
		},
		{
			Name:          "Test API 2",
			Label:         "Test Label 2",
			IsDefault:     false,
			Url:           "http://test.url2.com",
			ApiAuthHeader: "Authorization",
			ApiAuthKey:    "TestKey2",
			UserID:        admin.ID,
		},
	}

	for _, api := range expectedResults {
		_, err := q.CreateChatModel(context.Background(), sqlc_queries.CreateChatModelParams{
			Name:          api.Name,
			Label:         api.Label,
			IsDefault:     api.IsDefault,
			Url:           api.Url,
			ApiAuthHeader: api.ApiAuthHeader,
			ApiAuthKey:    api.ApiAuthKey,
			UserID:        api.UserID,
		})
		if err != nil {
			fmt.Printf("Error creating test data: %s", err.Error())
		}
	}
	return admin, expectedResults
}
func clearChatModelsIfExists(q *sqlc_queries.Queries) {
	defaultApis, _ := q.ListChatModels(context.Background())

	for _, api := range defaultApis {
		q.DeleteChatModel(context.Background(),
			sqlc_queries.DeleteChatModelParams{
				ID:     api.ID,
				UserID: api.UserID,
			})
	}
}

func unmarshalResponseToChatModel(t *testing.T, rr *httptest.ResponseRecorder) []sqlc_queries.ChatModel {
	// read the response body
	// unmarshal the response body into a list of ChatModel
	var results []sqlc_queries.ChatModel
	err := json.NewDecoder(rr.Body).Decode(&results)
	assert.NilError(t, err)

	return results
}

// the code below do db update directly in instead of using handler, please change to use handler
func TestChatModel(t *testing.T) {
	q := sqlc_queries.New(db)
	h := NewChatModelHandler(q) // create a new ChatModelHandler instance for testing
	router := mux.NewRouter()
	h.Register(router)
	// delete all existing chat APIs
	clearChatModelsIfExists(q)

	// Now let's create our expected results. Create two results and insert them into the database using the queries.
	admin, expectedResults := createTwoChatModel(q)

	// ensure that we get an array of two chat APIs in the response body
	// ensure the returned values are what we expect them to be
	results := checkGetModels(t, router, expectedResults)

	// Now lets update the the first element of our expected results array and call PUT on the endpoint

	// Create an HTTP request so we can simulate a PUT with the payload
	// ensure the new values are returned and were also updated in the database
	firstRecordID := results[0].ID
	updateFirstRecord(t, router, firstRecordID, admin, expectedResults[0])

	// delete first model
	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/chat_model/%d", firstRecordID), nil)
	deleteReq = deleteReq.WithContext(getContextWithUser(int(admin.ID)))
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)
	assert.Equal(t, deleteRR.Code, http.StatusOK)

	// check only one model left
	req, _ := http.NewRequest("GET", "/chat_model", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// ensure that we get an array of one chat API in the response body
	results = unmarshalResponseToChatModel(t, rr)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0].Name, "Test API 1")

	// delete the last model
	deleteRequest, _ := http.NewRequest("DELETE", fmt.Sprintf("/chat_model/%d", results[0].ID), nil)
	contextWithUser := getContextWithUser(int(admin.ID))
	deleteRequest = deleteRequest.WithContext(contextWithUser)
	deleteResponseRecorder := httptest.NewRecorder()
	router.ServeHTTP(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, deleteResponseRecorder.Code, http.StatusOK)

	// check no models left
	getRequest, _ := http.NewRequest("GET", "/chat_model", nil)
	// Create a ResponseRecorder to record the response
	getResponseRecorder := httptest.NewRecorder()
	router.ServeHTTP(getResponseRecorder, getRequest)
	results = unmarshalResponseToChatModel(t, getResponseRecorder)
	assert.Equal(t, len(results), 0)
}

func checkGetModels(t *testing.T, router *mux.Router, expectedResults []sqlc_queries.ChatModel) []sqlc_queries.ChatModel {
	req, _ := http.NewRequest("GET", "/chat_model", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	var results []sqlc_queries.ChatModel
	err := json.NewDecoder(rr.Body).Decode(&results)
	if err != nil {
		t.Errorf("error parsing response body: %s", err.Error())
	}
	assert.Equal(t, len(results), 2)
	assert.DeepEqual(t, lo.Reverse(expectedResults), results, cmpopts.IgnoreFields(sqlc_queries.ChatModel{}, "ID"))
	return results
}

func updateFirstRecord(t *testing.T, router *mux.Router, chatModelID int32, admin sqlc_queries.AuthUser, rec sqlc_queries.ChatModel) {
	rec.Name = "Test API 1 Updated"
	rec.Label = "Test Label 1 Updated"

	updateBytes, err := json.Marshal(rec)
	if err != nil {
		t.Errorf("Error marshaling update payload: %s", err.Error())
	}

	updateReq, _ := http.NewRequest("PUT", fmt.Sprintf("/chat_model/%d", chatModelID), bytes.NewBuffer(updateBytes))
	updateReq = updateReq.WithContext(getContextWithUser(int(admin.ID)))

	updateRR := httptest.NewRecorder()

	router.ServeHTTP(updateRR, updateReq)

	assert.Equal(t, updateRR.Code, http.StatusOK)

	var updatedResult sqlc_queries.ChatModel
	err = json.Unmarshal(updateRR.Body.Bytes(), &updatedResult)
	if err != nil {
		t.Errorf("Error parsing response body: %s", err.Error())
	}

	assert.Equal(t, rec.Name, updatedResult.Name)
	assert.Equal(t, rec.Label, updatedResult.Label)
}
