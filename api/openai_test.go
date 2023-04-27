package main

import "testing"

func Test_getModelBaseUrl(t *testing.T) {

	testCases := []struct {
		name     string
		apiUrl   string
		expected string
	}{
		{
			name:     "Base URL with v1 version",
			apiUrl:   "https://api.openai-sb.com/v1/chat/completions",
			expected: "https://api.openai-sb.com/v1",
		},
		{
			name:     "Base URL with v2 version",
			apiUrl:   "https://api.openai-sb.com/v2/completions",
			expected: "https://api.openai-sb.com/v2",
		},
		{
			name:     "Base URL with no version",
			apiUrl:   "https://api.openai-sb.com/chat/completions",
			expected: "https://api.openai-sb.com/chat",
		},
		{
			name:     "Base URL with different host",
			apiUrl:   "https://example.com/v1/chat/completions",
			expected: "https://example.com/v1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := getModelBaseUrl(tc.apiUrl)
			if actual != tc.expected {
				t.Errorf("Expected base URL '%s', but got '%s'", tc.expected, actual)
			}
		})
	}
}
