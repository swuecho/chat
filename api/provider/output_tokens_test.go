package provider

import "testing"

func TestSessionOutputTokens(t *testing.T) {
	if got := (Session{}).OutputTokens(); got != DefaultOutputTokens {
		t.Fatalf("unset MaxTokens: got %d, want %d", got, DefaultOutputTokens)
	}
	if got := (Session{MaxTokens: 64}).OutputTokens(); got != 64 {
		t.Fatalf("explicit MaxTokens: got %d, want 64", got)
	}
}
