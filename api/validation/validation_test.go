package validation

import (
	"strings"
	"testing"
)

func TestUUID(t *testing.T) {
	if err := UUID("sessionUuid", "01990a45-8a36-7e51-bf7c-a8df8d6b8e91", true); err != nil {
		t.Fatalf("valid UUID rejected: %v", err)
	}
	if err := UUID("sessionUuid", "not-a-uuid", true); err == nil {
		t.Fatal("invalid UUID accepted")
	}
}

func TestTextLimits(t *testing.T) {
	if err := Topic(strings.Repeat("界", MaxTopicLength), true); err != nil {
		t.Fatalf("topic at boundary rejected: %v", err)
	}
	if err := Topic(strings.Repeat("界", MaxTopicLength+1), true); err == nil {
		t.Fatal("oversized topic accepted")
	}
	if err := ModelName("model", "bad\nmodel", true); err == nil {
		t.Fatal("model containing control characters accepted")
	}
}

func TestTokenCount(t *testing.T) {
	if err := TokenCount("maxTokens", 1, false); err != nil {
		t.Fatalf("minimum token count rejected: %v", err)
	}
	if err := TokenCount("maxTokens", MaxTokenCount+1, false); err == nil {
		t.Fatal("oversized token count accepted")
	}
}
