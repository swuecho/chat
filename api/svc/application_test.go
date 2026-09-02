package svc

import (
	"testing"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestApplicationServicesSharesCoreDependencies(t *testing.T) {
	q := sqlc_queries.New(nil)
	app := NewApplicationServices(q, "key", "proxy", "secret", 100)
	if app.Chat == nil || app.Sessions == nil || app.Conversations == nil || app.Messages == nil {
		t.Fatal("core application services were not composed")
	}
	if app.Chat.q != q || app.Sessions.q != q || app.Conversations.q != q || app.Messages.q != q {
		t.Fatal("application services do not share the composition-root query dependency")
	}
	if app.Chat.SuggestionGenerator() == nil || app.Chat.AuditLogger() == nil {
		t.Fatal("chat supporting components were not composed")
	}
	if app.ChatUseCases == nil || app.ChatUseCases.chat != app.Chat || app.ChatUseCases.sessions != app.Sessions {
		t.Fatal("chat use-case factory does not share the application services")
	}
}
