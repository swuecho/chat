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

// SummarizeWithTimeout generates a summary with a 20-second timeout.
func SummarizeWithTimeout(apiToken, baseURL, content string) string {
	// Create a context with a 20 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Call the summarize function with the context
	summary := llm_summarize(ctx, apiToken, baseURL, content)

	return summary
}

func llm_summarize(ctx context.Context, apiToken, baseURL string, doc string) string {
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	llm, err := openai.New(
		openai.WithToken(apiToken),
		openai.WithBaseURL(baseURL),
	)
	if err != nil {
		slog.Info("failed to create openai client", "baseURL", baseURL, "error", err)
		return ""
	}

	llmSummarizationChain := chains.LoadRefineSummarization(llm)
	docs, _ := documentloaders.NewText(strings.NewReader(doc)).LoadAndSplit(ctx,
		textsplitter.NewRecursiveCharacter(),
	)
	outputValues, err := chains.Call(ctx, llmSummarizationChain, map[string]any{"input_documents": docs})
	if err != nil {
		slog.Info("failed to call chain", "baseURL", baseURL, "error", err)
		return ""
	}
	out, _ := outputValues["text"].(string)
	return out
}
