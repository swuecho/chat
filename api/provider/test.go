package provider

import (
	"context"

	"github.com/swuecho/chat_backend/models"
)

// TestChatModel implements ChatModel for testing.
type TestChatModel struct {
	h Handler
}

// NewTestChatModel creates a new TestChatModel.
func NewTestChatModel(h Handler) *TestChatModel {
	return &TestChatModel{h: h}
}

func (m *TestChatModel) Stream(ctx context.Context, input Request) (<-chan StreamChunk, error) {
	chatUuid, regenerate := input.ChatUUID, input.Regenerate

	answerID := generateAnswerID(chatUuid, regenerate, input.NewID)
	answer := "Hi, I am a chatbot. I can help you to find the best answer for your question. Please ask me a question."

	ch := make(chan StreamChunk, 2)
	go func() {
		defer close(ch)
		if !emitChunk(ctx, ch, StreamChunk{ID: answerID, Content: answer}) {
			return
		}

		emitChunk(ctx, ch, StreamChunk{
			ID:          answerID,
			Done:        true,
			FinalAnswer: &models.LLMAnswer{Answer: answer, AnswerId: answerID},
		})
	}()
	return ch, nil
}
