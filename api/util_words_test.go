package main

import (
	"testing"
)

func Test_firstNWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected string
	}{
		{
			name:     "normal case with exactly 10 words",
			input:    "how do I write a function that processes data efficiently in Go",
			n:        10,
			expected: "how do I write a function that processes data efficiently",
		},
		{
			name:     "less than 10 words",
			input:    "hello world how are you",
			n:        10,
			expected: "hello world how are you",
		},
		{
			name:     "more than 10 words",
			input:    "this is a very long prompt that contains much more than ten words so it should be truncated",
			n:        10,
			expected: "this is a very long prompt that contains much more",
		},
		{
			name:     "empty string",
			input:    "",
			n:        10,
			expected: "",
		},
		{
			name:     "single word",
			input:    "hello",
			n:        10,
			expected: "hello",
		},
		{
			name:     "exactly n words",
			input:    "one two three four five",
			n:        5,
			expected: "one two three four five",
		},
		{
			name:     "with extra whitespace",
			input:    "  hello   world   how   are   you  ",
			n:        3,
			expected: "hello world how",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := firstNWords(tt.input, tt.n)
			if result != tt.expected {
				t.Errorf("firstNWords(%q, %d) = %q, want %q", tt.input, tt.n, result, tt.expected)
			}
		})
	}
}
