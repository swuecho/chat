package main

import "testing"

func TestEmbedInstructions(t *testing.T) {
	if artifactInstructionText == "" {
		t.Fatalf("artifactInstructionText is empty")
	}
}
