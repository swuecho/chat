package main

import (
	"testing"

	"github.com/swuecho/chat_backend/svc"
)

func TestEmbedInstructions(t *testing.T) {
	instructions, err := svc.LoadArtifactInstruction()
	if err != nil {
		t.Skip("artifact instruction not available:", err)
	}
	if instructions == "" {
		t.Fatal("artifact instruction text is empty")
	}
}
