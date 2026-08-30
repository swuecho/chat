package util

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

type validatingRequest struct {
	Name string `json:"name"`
}

func (r *validatingRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "valid", body: `{"name":"chat"}`},
		{name: "empty", body: ``, wantErr: "must not be empty"},
		{name: "unknown field", body: `{"name":"chat","extra":true}`, wantErr: "unknown field"},
		{name: "trailing value", body: `{"name":"chat"} {"name":"again"}`, wantErr: "exactly one JSON value"},
		{name: "invalid syntax", body: `{"name":`, wantErr: "decode request body"},
		{name: "validation", body: `{"name":" "}`, wantErr: "name is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			var target validatingRequest
			err := DecodeJSON(req, &target)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("DecodeJSON() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("DecodeJSON() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
