package main

import (
	"os"
	"testing"
)

// TestMain is the test entry point for the main package.
// Tests here (util_test.go, util_words_test.go, embed_debug_test.go)
// do not need a database connection.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
