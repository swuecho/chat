package svc

import (
	"regexp"
	"strings"

	"github.com/swuecho/chat_backend/domain"
)

// extractArtifacts detects and extracts artifacts from message content
func extractArtifacts(content string, newID func() string) []domain.Artifact {
	var artifacts []domain.Artifact

	// Pattern for HTML artifacts (check specific types first)
	// Example: ```html <!-- artifact: Interactive Demo -->
	htmlArtifactRegex := regexp.MustCompile(`(?is)` + "```" + `html\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n` + "```")
	htmlMatches := htmlArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range htmlMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])
		if len(title) > maxArtifactTitleBytes || len(artifactContent) > maxArtifactContentBytes {
			continue
		}
		artifact := domain.Artifact{
			UUID:     newID(),
			Type:     "html",
			Title:    title,
			Content:  artifactContent,
			Language: "html",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for SVG artifacts
	// Example: ```svg <!-- artifact: Logo Design -->
	svgArtifactRegex := regexp.MustCompile(`(?is)` + "```" + `svg\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n` + "```")
	svgMatches := svgArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range svgMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		if len(title) > maxArtifactTitleBytes || len(artifactContent) > maxArtifactContentBytes {
			continue
		}
		artifact := domain.Artifact{
			UUID:     newID(),
			Type:     "svg",
			Title:    title,
			Content:  artifactContent,
			Language: "svg",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for Mermaid diagrams
	// Example: ```mermaid <!-- artifact: Flow Chart -->
	mermaidArtifactRegex := regexp.MustCompile(`(?is)` + "```" + `mermaid\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n` + "```")
	mermaidMatches := mermaidArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range mermaidMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		if len(title) > maxArtifactTitleBytes || len(artifactContent) > maxArtifactContentBytes {
			continue
		}
		artifact := domain.Artifact{
			UUID:     newID(),
			Type:     "mermaid",
			Title:    title,
			Content:  artifactContent,
			Language: "mermaid",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for JSON artifacts
	// Example: ```json <!-- artifact: API Response -->
	jsonArtifactRegex := regexp.MustCompile(`(?is)` + "```" + `json\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n` + "```")
	jsonMatches := jsonArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range jsonMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		if len(title) > maxArtifactTitleBytes || len(artifactContent) > maxArtifactContentBytes {
			continue
		}
		artifact := domain.Artifact{
			UUID:     newID(),
			Type:     "json",
			Title:    title,
			Content:  artifactContent,
			Language: "json",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for general code artifacts (exclude html and svg which are handled above)
	// Example: ```javascript <!-- artifact: React Component -->
	codeArtifactRegex := regexp.MustCompile(`(?is)` + "```" + `([\w+-]+)?\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n` + "```")
	matches := codeArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		language := strings.ToLower(match[1])
		title := strings.TrimSpace(match[2])
		artifactContent := strings.TrimSpace(match[3])

		// Skip formats handled by their specialized renderers.
		if language == "html" || language == "svg" || language == "mermaid" || language == "json" {
			continue
		}

		if language == "" {
			language = "text"
		}

		if len(title) > maxArtifactTitleBytes || len(artifactContent) > maxArtifactContentBytes {
			continue
		}
		artifact := domain.Artifact{
			UUID:     newID(),
			Type:     "code",
			Title:    title,
			Content:  artifactContent,
			Language: language,
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts
}
