package provider

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// TestChatModel implements ChatModel for testing.
type TestChatModel struct {
	h Handler
}

// NewTestChatModel creates a new TestChatModel.
func NewTestChatModel(h Handler) *TestChatModel {
	return &TestChatModel{h: h}
}

func (m *TestChatModel) Stream(w http.ResponseWriter, session sqlc_queries.ChatSession,
	messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {

	chatFiles, err := GetChatFiles(m.h.RequestContext(), m.h.Queries(), session.Uuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to get chat files").WithDebugInfo(err.Error()))
		return nil, err
	}

	answerID := generateAnswerID(chatUuid, regenerate)
	flusher, err := SetupSSEStream(w)
	if err != nil {
		dto.RespondWithAPIError(w, dto.APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		})
		return nil, err
	}

	answer := "Hi, I am a chatbot. I can help you to find the best answer for your question. Please ask me a question."
	if err := FlushResponse(w, flusher, StreamingResponse{AnswerID: answerID, Content: answer}); err != nil {
		slog.Info("Failed to flush response: %v", err)
	}

	if session.Debug {
		openaiReq := NewChatCompletionRequest(session, messages, chatFiles, false)
		reqJ, _ := json.Marshal(openaiReq)
		FlushResponse(w, flusher, StreamingResponse{AnswerID: answerID, Content: answer + "\n" + string(reqJ), IsFinal: true})
	}

	return &models.LLMAnswer{Answer: answer, AnswerId: answerID}, nil
}
