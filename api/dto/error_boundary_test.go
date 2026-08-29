package dto

import (
	"errors"
	"net/http"
	"testing"

	"github.com/swuecho/chat_backend/domain"
)

func TestToAPIErrorMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"invalid", domain.Invalid("bad input"), http.StatusBadRequest, ErrValidationInvalidInputGeneric.Code},
		{"unauthorized", domain.Unauthorized("sign in required"), http.StatusUnauthorized, ErrAuthInvalidCredentials.Code},
		{"forbidden", domain.Forbidden("access denied"), http.StatusForbidden, ErrAuthAccessDenied.Code},
		{"not found", domain.NotFound("Chat session", errors.New("missing row")), http.StatusNotFound, ErrChatSessionNotFound.Code},
		{"generic not found", domain.NotFound("Workspace", errors.New("missing row")), http.StatusNotFound, ErrResourceNotFoundGeneric.Code},
		{"conflict", domain.Conflict("already exists", errors.New("duplicate")), http.StatusConflict, ErrResourceAlreadyExistsGeneric.Code},
		{"unavailable", domain.Unavailable("provider unavailable", errors.New("offline")), http.StatusServiceUnavailable, ErrExternalUnavailable.Code},
		{"internal", domain.Internal("operation failed", errors.New("boom")), http.StatusInternalServerError, ErrInternalUnexpected.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToAPIError(tt.err)
			if got.HTTPCode != tt.status || got.Code != tt.code {
				t.Fatalf("got status/code %d/%s, want %d/%s", got.HTTPCode, got.Code, tt.status, tt.code)
			}
		})
	}
}
