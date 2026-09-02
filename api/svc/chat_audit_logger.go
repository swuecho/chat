package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatAuditLogger interface {
	LogChat(context.Context, ChatSession, []models.Message, string) error
}

type SQLChatAuditLogger struct{ q *sqlc_queries.Queries }

func NewSQLChatAuditLogger(q *sqlc_queries.Queries) *SQLChatAuditLogger {
	return &SQLChatAuditLogger{q: q}
}

// logChat creates a chat log entry for analytics and debugging.
// Logs the session, messages, and LLM response for audit purposes.
func (s *SQLChatAuditLogger) LogChat(ctx context.Context, chatSession ChatSession, msgs []models.Message, answerText string) error {
	// log chat
	sessionRaw := chatSession.ToRawMessage()
	if sessionRaw == nil {
		slog.Info("failed to marshal chat session")
		return fmt.Errorf("marshal chat session")
	}
	question, err := json.Marshal(msgs)
	if err != nil {
		slog.Warn("Failed to marshal chat messages", "error", err)
		return err
	}
	answerRaw, err := json.Marshal(answerText)
	if err != nil {
		slog.Warn("Failed to marshal answer", "error", err)
		return err
	}

	_, err = s.q.CreateChatLog(ctx, sqlc_queries.CreateChatLogParams{
		Session:  *sessionRaw,
		Question: question,
		Answer:   answerRaw,
	})
	return err
}
