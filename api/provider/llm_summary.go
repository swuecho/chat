package provider

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
)

// OpenAIToken is set during initialization to avoid importing main package globals.
var OpenAIToken string

// SummarizeWithTimeout generates a summary with a 20-second timeout.
func SummarizeWithTimeout(baseURL, content string) string {
	// Create a context with a 20 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Call the summarize function with the context
	summary := llm_summarize(ctx, baseURL, content)

	return summary
}

func llm_summarize(ctx context.Context, baseURL string, doc string) string {
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	llm, err := openai.New(
		openai.WithToken(OpenAIToken),
		openai.WithBaseURL(baseURL),
	)
	if err != nil {
		slog.Info("failed to create openai client %s: %v", baseURL, err)
		return ""
	}

	llmSummarizationChain := chains.LoadRefineSummarization(llm)
	docs, _ := documentloaders.NewText(strings.NewReader(doc)).LoadAndSplit(ctx,
		textsplitter.NewRecursiveCharacter(),
	)
	outputValues, err := chains.Call(ctx, llmSummarizationChain, map[string]any{"input_documents": docs})
	if err != nil {
		slog.Info("failed to call chain: %s, %v", baseURL, err)
		return ""
	}
	out, _ := outputValues["text"].(string)
	return out
}
