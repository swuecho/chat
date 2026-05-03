package main

import (
	"testing"

	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func Test_getModelBaseUrl(t *testing.T) {

	testCases := []struct {
		name     string
		apiUrl   string
		expected string
	}{
		{
			name:     "Base URL with different host",
			apiUrl:   "https://example.com/v1/chat/completions",
			expected: "https://example.com/v1",
		},
		{
			name:     "Base URL with different host",
			apiUrl:   "https://docs-test-001.openai.azure.com/",
			expected: "https://docs-test-001.openai.azure.com/",
		},
		{
			name:     "BigModel base URL without endpoint suffix",
			apiUrl:   "https://open.bigmodel.cn/api/paas/v4",
			expected: "https://open.bigmodel.cn/api/paas/v4",
		},
		{
			name:     "BigModel full chat completions endpoint",
			apiUrl:   "https://open.bigmodel.cn/api/paas/v4/chat/completions",
			expected: "https://open.bigmodel.cn/api/paas/v4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := provider.GetModelBaseURL(tc.apiUrl)
			if actual != tc.expected {
				t.Errorf("Expected base URL '%s', but got '%s'", tc.expected, actual)
			}
		})
	}
}

func Test_normalizeOpenAIModelName(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		modelName string
		expected  string
	}{
		{
			name:      "BigModel lowercases model name",
			url:       "https://open.bigmodel.cn/api/paas/v4",
			modelName: "GLM-5.1",
			expected:  "glm-5.1",
		},
		{
			name:      "Other providers keep model name",
			url:       "https://api.openai.com/v1",
			modelName: "GPT-4o",
			expected:  "GPT-4o",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := provider.NormalizeOpenAIModelName(sqlc_queries.ChatModel{Url: tc.url}, tc.modelName)
			if actual != tc.expected {
				t.Errorf("Expected model name '%s', but got '%s'", tc.expected, actual)
			}
		})
	}
}
