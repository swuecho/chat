package gemini

import (
	"os"
	"testing"
)

func TestBuildAPIURL(t *testing.T) {
	// Set test environment variable
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	tests := []struct {
		name     string
		model    string
		stream   bool
		expected string
	}{
		{
			name:   "non-streaming request",
			model:  "gemini-pro",
			stream: false,
			expected: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:" +
				"generateContent?key=test-key",
		},
		{
			name:   "streaming request",
			model:  "gemini-pro",
			stream: true,
			expected: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:" +
				"streamGenerateContent?alt=sse&key=test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAPIURL(tt.model, tt.stream)
			if got != tt.expected {
				t.Errorf("buildAPIURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
