package dto

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIError(t *testing.T) {
	err := ErrResourceNotFound("Session")
	if err.HTTPCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.HTTPCode)
	}
	if err.Message != "Session not found" {
		t.Errorf("expected 'Session not found', got %q", err.Message)
	}
}

func TestWrapError(t *testing.T) {
	original := errors.New("db down")
	wrapped := WrapError(original, "during query")
	if wrapped.HTTPCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", wrapped.HTTPCode)
	}
	if wrapped.DebugInfo != "db down" {
		t.Errorf("expected debug info, got %q", wrapped.DebugInfo)
	}
}

func TestWrapAPIError(t *testing.T) {
	original := ErrAuthInvalidCredentials
	wrapped := WrapError(original, "login handler")
	if wrapped.Code != ErrAuth+"_001" {
		t.Errorf("expected AUTH_001, got %s", wrapped.Code)
	}
}

func TestErrorCatalogHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/errors", nil)
	w := httptest.NewRecorder()
	ErrorCatalogHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSimpleChatMessageGetRole(t *testing.T) {
	user := SimpleChatMessage{Inversion: true}
	if user.GetRole() != "user" {
		t.Errorf("expected 'user', got %q", user.GetRole())
	}
	assistant := SimpleChatMessage{Inversion: false}
	if assistant.GetRole() != "assistant" {
		t.Errorf("expected 'assistant', got %q", assistant.GetRole())
	}
}

func TestPaginationOffset(t *testing.T) {
	p := Pagination{Page: 3, Size: 20}
	if p.Offset() != 40 {
		t.Errorf("expected 40, got %d", p.Offset())
	}
}
