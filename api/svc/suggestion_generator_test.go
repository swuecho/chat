package svc

import (
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestSelectSuggestionModel(t *testing.T) {
	claude := sqlc_queries.ChatModel{ID: 1, Name: "claude-sonnet", ApiType: "claude", IsEnable: true}
	gpt := sqlc_queries.ChatModel{ID: 2, Name: "gpt-4o", ApiType: "openai", IsEnable: true}
	flash := sqlc_queries.ChatModel{ID: 3, Name: "gemini-2.0-flash", ApiType: "gemini", IsEnable: true}
	disabledFlash := sqlc_queries.ChatModel{ID: 4, Name: "deepseek-chat", ApiType: "openai", IsEnable: false}

	cases := []struct {
		name   string
		models []sqlc_queries.ChatModel
		wantID int32
		wantOK bool
	}{
		{"preferred model wins", []sqlc_queries.ChatModel{claude, gpt, flash}, 3, true},
		{"disabled preferred model is skipped", []sqlc_queries.ChatModel{claude, disabledFlash, gpt}, 2, true},
		{"falls back to any enabled model", []sqlc_queries.ChatModel{disabledFlash, claude}, 1, true},
		{"no enabled models", []sqlc_queries.ChatModel{disabledFlash}, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := selectSuggestionModel(tc.models)
			if ok != tc.wantOK || got.ID != tc.wantID {
				t.Fatalf("selectSuggestionModel() = (%d, %v), want (%d, %v)", got.ID, ok, tc.wantID, tc.wantOK)
			}
		})
	}
}
