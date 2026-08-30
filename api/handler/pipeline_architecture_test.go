package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Ordinary handlers have one response/error boundary. The listed files are
// explicit protocol boundaries and are reviewed separately for SSE, proxy, or
// binary response behavior.
func TestOrdinaryHandlersDoNotWriteAdHocResponses(t *testing.T) {
	protocolFiles := map[string]bool{
		"chat_session_helpers.go": true,
		"chat_stream.go":          true,
		"gateway.go":              true,
		"tts.go":                  true,
		"util.go":                 true,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || protocolFiles[name] {
			continue
		}
		contents, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatal(err)
		}
		source := string(contents)
		for _, forbidden := range []string{"dto.RespondWithAPIError(", "json.NewEncoder(w)", "w.WriteHeader("} {
			if strings.Contains(source, forbidden) {
				t.Errorf("%s bypasses the ordinary endpoint response boundary with %q", name, forbidden)
			}
		}
	}
}
