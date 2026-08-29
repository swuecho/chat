package svc

import "testing"

func TestExtractArtifactsUsesInjectedIDGenerator(t *testing.T) {
	nextID := func() string { return "artifact-id" }
	artifacts := extractArtifacts("```html <!-- artifact: Demo -->\n<h1>Hello</h1>\n```", nextID)

	if len(artifacts) != 1 {
		t.Fatalf("expected one artifact, got %d", len(artifacts))
	}
	if artifacts[0].UUID != "artifact-id" {
		t.Fatalf("expected injected ID, got %q", artifacts[0].UUID)
	}
	if artifacts[0].Type != "html" || artifacts[0].Title != "Demo" {
		t.Fatalf("unexpected artifact: %#v", artifacts[0])
	}
}
