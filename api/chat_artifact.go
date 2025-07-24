package main

import (
	"regexp"
	"strings"
)

// extractArtifacts detects and extracts artifacts from message content
func extractArtifacts(content string) []Artifact {
	var artifacts []Artifact

	// Pattern for HTML artifacts (check specific types first)
	// Example: ```html <!-- artifact: Interactive Demo -->
	htmlArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `html\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	htmlMatches := htmlArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range htmlMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])
		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "html",
			Title:    title,
			Content:  artifactContent,
			Language: "html",
		}
		artifacts = append(artifacts, artifact)
	}


	// Pattern for SVG artifacts
	// Example: ```svg <!-- artifact: Logo Design -->
	svgArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `svg\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	svgMatches := svgArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range svgMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "svg",
			Title:    title,
			Content:  artifactContent,
			Language: "svg",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for Mermaid diagrams
	// Example: ```mermaid <!-- artifact: Flow Chart -->
	mermaidArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `mermaid\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	mermaidMatches := mermaidArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range mermaidMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "mermaid",
			Title:    title,
			Content:  artifactContent,
			Language: "mermaid",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for JSON artifacts
	// Example: ```json <!-- artifact: API Response -->
	jsonArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `json\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	jsonMatches := jsonArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range jsonMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "json",
			Title:    title,
			Content:  artifactContent,
			Language: "json",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for executable code artifacts
	// Example: ```javascript <!-- executable: Calculator -->
	executableArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `(\w+)?\s*<!--\s*executable:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	executableMatches := executableArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range executableMatches {
		language := match[1]
		title := strings.TrimSpace(match[2])
		artifactContent := strings.TrimSpace(match[3])

		// Skip if already processed as HTML, SVG, Mermaid, or JSON
		if language == "html" || language == "svg" || language == "mermaid" || language == "json" {
			continue
		}

		if language == "" {
			language = "javascript" // Default to JavaScript for executable code
		}

		// Only create executable artifacts for supported languages
		if isExecutableLanguage(language) {
			artifact := Artifact{
				UUID:     NewUUID(),
				Type:     "executable-code",
				Title:    title,
				Content:  artifactContent,
				Language: language,
			}
			artifacts = append(artifacts, artifact)
		}
	}

	// Pattern for general code artifacts (exclude html and svg which are handled above)
	// Example: ```javascript <!-- artifact: React Component -->
	codeArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `(\w+)?\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	matches := codeArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		language := match[1]
		title := strings.TrimSpace(match[2])
		artifactContent := strings.TrimSpace(match[3])

		// Skip if already processed as HTML, SVG, Mermaid, JSON, or executable
		if language == "html" || language == "svg" || language == "mermaid" || language == "json" {
			continue
		}

		if language == "" {
			language = "text"
		}

		// Check if this should be an executable artifact for supported languages
		artifactType := "code"
		if isExecutableLanguage(language) {
			// For supported languages, make them executable by default if they contain certain patterns
			if containsExecutablePatterns(artifactContent) {
				artifactType = "executable-code"
			}
		}

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     artifactType,
			Title:    title,
			Content:  artifactContent,
			Language: language,
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts
}

// isExecutableLanguage checks if a language is supported for code execution
func isExecutableLanguage(language string) bool {
	executableLanguages := []string{
		"javascript", "js", "typescript", "ts",
		"python", "py",
	}
	
	language = strings.ToLower(strings.TrimSpace(language))
	for _, execLang := range executableLanguages {
		if language == execLang {
			return true
		}
	}
	return false
}

// containsExecutablePatterns checks if code contains patterns that suggest it should be executable
func containsExecutablePatterns(content string) bool {
	// Patterns that suggest the code is meant to be executed
	executablePatterns := []string{

		// JavaScript patterns
		"console.log",
		"console.error",
		"console.warn",
		"function",
		"const ",
		"let ",
		"var ",
		"=>",
		"if (",
		"for (",
		"while (",
		"return ",

		// Python patterns
		"print(",
		"import ",
		"from ",
		"def ",
		"if __name__",
		"class ",
		"for ",
		"while ",

	}
	
	contentLower := strings.ToLower(content)
	for _, pattern := range executablePatterns {
		if strings.Contains(contentLower, pattern) {
			return true
		}
	}
	return false
}
