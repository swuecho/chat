package main

import "testing"

func TestEmbedInstructions(t *testing.T) {
	if artifactInstructionText == "" {
		t.Fatalf("artifactInstructionText is empty")
	}
	if toolInstructionText == "" {
		t.Fatalf("toolInstructionText is empty")
	}
}
