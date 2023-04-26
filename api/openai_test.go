package main

import (
	"testing"
)

func TestExtractBaseURL(t *testing.T) {
	testCases := []struct {
		name        string
		rawURL      string
		wantBaseURL string
		wantErr     bool
	}{
		{
			name:        "valid: has version",
			rawURL:      "https://api.openai.com/v1/chat/completions",
			wantBaseURL: "https://api.openai.com/v1/",
			wantErr:     false,
		},
		{
			name:        "valid: no version",
			rawURL:      "https://api.myapp.com/query",
			wantBaseURL: "https://api.myapp.com/",
			wantErr:     false,
		},
		{
			name:        "invalid: empty URL",
			rawURL:      "",
			wantBaseURL: "",
			wantErr:     true,
		},
		{
			name:        "invalid: not a URL",
			rawURL:      "not a url",
			wantBaseURL: "",
			wantErr:     true,
		},
		{
			name:        "invalid: URL scheme missing",
			rawURL:      "myapp.com",
			wantBaseURL: "",
			wantErr:     true,
		},
		{
			name:        "invalid: URL host missing",
			rawURL:      "https:///query",
			wantBaseURL: "",
			wantErr:     true,
		},
		{
			name:        "invalid: version missing",
			rawURL:      "https://api.myapp.com/",
			wantBaseURL: "",
			wantErr:     true,
		},
		{
			name:        "v2",
			rawURL:      "https://api.myapp.com/v2/query",
			wantBaseURL: "https://api.myapp.com/v2",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseURL, err := getModelBaseUrl(tc.rawURL)
			if (err != nil) != tc.wantErr {
				t.Errorf("Unexpected error status: %v", err)
			}
			if baseURL != tc.wantBaseURL {
				t.Errorf("Wrong base URL: got %q, want %q", baseURL, tc.wantBaseURL)
			}
		})
	}
}
