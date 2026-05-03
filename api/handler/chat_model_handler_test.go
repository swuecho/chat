package handler

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
			Name: "Test API 1", Label: "Test Label 1", IsDefault: false,
			Url: "http://test.url.com", ApiAuthHeader: "Authorization",
			ApiAuthKey: "TestKey1", UserID: admin.ID,
		},
		{
			Name: "Test API 2", Label: "Test Label 2", IsDefault: false,
			Url: "http://test.url2.com", ApiAuthHeader: "Authorization",
			ApiAuthKey: "TestKey2", UserID: admin.ID,
		},
	}
	for _, api := range expectedResults {
		if _, err := q.CreateChatModel(context.Background(), sqlc_queries.CreateChatModelParams{
			Name:          api.Name,
			Label:         api.Label,
			IsDefault:     api.IsDefault,
			Url:           api.Url,
			ApiAuthHeader: api.ApiAuthHeader,
			ApiAuthKey:    api.ApiAuthKey,
			UserID:        api.UserID,
		}); err != nil {
			fmt.Printf("Error creating test data: %s", err.Error())
		}
	}
	return admin, expectedResults
}

func clearChatModelsIfExists(q *sqlc_queries.Queries) {
	defaultApis, _ := q.ListChatModels(context.Background())
	for _, api := range defaultApis {
		q.DeleteChatModel(context.Background(),
			sqlc_queries.DeleteChatModelParams{ID: api.ID, UserID: api.UserID})
	}
}

func unmarshalResponseToChatModel(t *testing.T, rr *httptest.ResponseRecorder) []sqlc_queries.ChatModel {
	var results []sqlc_queries.ChatModel
	err := json.NewDecoder(rr.Body).Decode(&results)
	assert.NilError(t, err)
	return results
}

func TestChatModelTest(t *testing.T) {
	q := sqlc_queries.New(testDB)
	h := NewChatModelHandler(q)
	router := mux.NewRouter()
	h.Register(router)
	clearChatModelsIfExists(q)

	admin, expectedResults := createTwoChatModel(q)
	results := checkGetModels(t, router, expectedResults)

	firstRecordID := results[0].ID
	updateFirstRecord(t, router, firstRecordID, admin, expectedResults[0])

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/chat_model/%d", firstRecordID), nil)
	deleteReq = deleteReq.WithContext(getContextWithUser(int(admin.ID)))
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)
	assert.Equal(t, deleteRR.Code, http.StatusOK)

	req, _ := http.NewRequest("GET", "/chat_model", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	results = unmarshalResponseToChatModel(t, rr)
	assert.Equal(t, len(results), 1)

	deleteReq2, _ := http.NewRequest("DELETE", fmt.Sprintf("/chat_model/%d", results[0].ID), nil)
	deleteReq2 = deleteReq2.WithContext(getContextWithUser(int(admin.ID)))
	deleteRR2 := httptest.NewRecorder()
	router.ServeHTTP(deleteRR2, deleteReq2)
	assert.Equal(t, deleteRR2.Code, http.StatusOK)

	getReq, _ := http.NewRequest("GET", "/chat_model", nil)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)
	results = unmarshalResponseToChatModel(t, getRR)
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
	assert.DeepEqual(t, lo.Reverse(expectedResults), results,
		cmpopts.IgnoreFields(sqlc_queries.ChatModel{}, "ID", "IsEnable"))
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
