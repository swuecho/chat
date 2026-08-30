package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/swuecho/chat_backend/models"
)

// ConsumeStream supervises one provider stream. It cancels the producer when
// delivery fails or a terminal value arrives and rejects malformed lifecycles.
func ConsumeStream(ctx context.Context, model ChatModel, request Request, onChunk func(StreamChunk) error) (*models.LLMAnswer, error) {
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := model.Stream(streamCtx, request)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, errors.New("provider returned a nil stream")
	}

	for {
		select {
		case <-streamCtx.Done():
			return nil, streamCtx.Err()
		case chunk, ok := <-ch:
			if !ok {
				if err := streamCtx.Err(); err != nil {
					return nil, err
				}
				return nil, errors.New("provider stream closed without a final answer")
			}
			if chunk.Err != nil {
				return nil, chunk.Err
			}
			if chunk.Done {
				if chunk.FinalAnswer == nil {
					return nil, errors.New("provider emitted an empty terminal chunk")
				}
				return chunk.FinalAnswer, nil
			}
			if onChunk != nil && chunk.Content != "" {
				if err := onChunk(chunk); err != nil {
					return nil, fmt.Errorf("deliver provider chunk: %w", err)
				}
			}
		}
	}
}
