package svc

import (
	"strings"
	"testing"
)

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

func TestExtractArtifactsIgnoresLegacyExecutableMarker(t *testing.T) {
	artifacts := extractArtifacts("```python <!-- executable: Report -->\nprint('hello')\n```", func() string { return "artifact-id" })

	if len(artifacts) != 0 {
		t.Fatalf("expected executable marker to be ignored, got %#v", artifacts)
	}
}

func TestExtractArtifactsAcceptsCaseInsensitiveLanguageAndCRLF(t *testing.T) {
	artifacts := extractArtifacts("```HTML <!-- artifact: Demo -->\r\n<p>Hello</p>\r\n```", func() string { return "artifact-id" })

	if len(artifacts) != 1 || artifacts[0].Type != "html" {
		t.Fatalf("expected HTML artifact, got %#v", artifacts)
	}
}

func TestExtractArtifactsSkipsOversizedContent(t *testing.T) {
	content := "```text <!-- artifact: Too large -->\n" + strings.Repeat("x", maxArtifactContentBytes+1) + "\n```"
	if artifacts := extractArtifacts(content, func() string { return "artifact-id" }); len(artifacts) != 0 {
		t.Fatalf("expected oversized artifact to be skipped, got %#v", artifacts)
	}
}
