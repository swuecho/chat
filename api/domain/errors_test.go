package domain

import (
	"errors"
	"testing"
)

func TestErrorPreservesKindAndCause(t *testing.T) {
	cause := errors.New("database failure")
	err := NotFound("Chat session", cause)

	if !IsKind(err, KindNotFound) {
		t.Fatalf("expected not-found error, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected application error to preserve its cause")
	}
}
