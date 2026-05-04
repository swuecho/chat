package main

import (
	"strings"
)

// firstN returns the first n characters (runes) of a string.
// Used by tests.
func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

// firstNWords returns the first n words of a string.
// Used by tests.
func firstNWords(s string, n int) string {
	if s == "" {
		return ""
	}
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ")
}
