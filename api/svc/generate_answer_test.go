package svc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type generateAnswerModel struct {
	chunks []provider.StreamChunk
}

type generateAnswerModelSelector struct{ model provider.ChatModel }

func (s generateAnswerModelSelector) SelectModel(context.Context, ChatSession, []models.Message) (provider.ChatModel, error) {
	return s.model, nil
}

type generateAnswerChunkSink struct{ delivered *string }

func (s generateAnswerChunkSink) WriteAnswerChunk(_ context.Context, chunk provider.StreamChunk) error {
	*s.delivered += chunk.Content
	return nil
}

type noChatLogs struct{}

func (noChatLogs) ShouldLogChat([]models.Message) bool { return false }

func (m generateAnswerModel) Stream(ctx context.Context, _ provider.Request) (<-chan provider.StreamChunk, error) {
	ch := make(chan provider.StreamChunk, len(m.chunks))
	go func() {
		defer close(ch)
		for _, chunk := range m.chunks {
			select {
			case <-ctx.Done():
				return
			case ch <- chunk:
			}
		}
	}()
	return ch, nil
}

func TestGenerateAnswerPersistsBeforeReturningSuccess(t *testing.T) {
	ctx := context.Background()
	q := sqlc_queries.New(testDB)
	user, err := q.CreateAuthUser(ctx, sqlc_queries.CreateAuthUserParams{
		Email: "generate-answer@test.com", Username: "generate-answer", Password: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteAuthUser(ctx, user.Email)

	model, err := q.CreateChatModel(ctx, sqlc_queries.CreateChatModelParams{
		Name: "generate-answer-model", Label: "Generate answer", Url: "http://provider.invalid/v1/chat/completions",
		UserID: user.ID, ApiType: "openai",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteChatModel(ctx, sqlc_queries.DeleteChatModelParams{ID: model.ID, UserID: user.ID})

	sessionRecord, err := q.CreateChatSession(ctx, sqlc_queries.CreateChatSessionParams{
		UserID: user.ID, Uuid: "generate-answer-session", Topic: "Generate answer", Model: model.Name,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams{
		Uuid: "generate-answer-prompt", ChatSessionUuid: sessionRecord.Uuid, Role: "system", Content: "Be useful",
		UserID: user.ID, CreatedBy: user.ID, UpdatedBy: user.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := q.CreateChatMessage(ctx, sqlc_queries.CreateChatMessageParams{
		Uuid: "generate-answer-question", ChatSessionUuid: sessionRecord.Uuid, Role: "user", Content: "Hello",
		Model: model.Name, UserID: user.ID, CreatedBy: user.ID, UpdatedBy: user.ID,
		Raw: json.RawMessage(`{}`), Artifacts: json.RawMessage(`[]`), SuggestedQuestions: json.RawMessage(`[]`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := q.ClaimChatRequest(ctx, sqlc_queries.ClaimChatRequestParams{
		Uuid: "generate-answer-request", ChatSessionUuid: sessionRecord.Uuid, UserID: user.ID,
	}); err != nil {
		t.Fatal(err)
	}

	service := NewChatService(q, "", "")
	var delivered string
	useCase := NewGenerateAnswerUseCase(service, generateAnswerModelSelector{model: generateAnswerModel{chunks: []provider.StreamChunk{
		{ID: "generate-answer-response", Content: "durable "},
		{ID: "generate-answer-response", Content: "answer"},
		{Done: true, FinalAnswer: &models.LLMAnswer{AnswerId: "generate-answer-response", Answer: "durable answer"}},
	}}}, generateAnswerChunkSink{delivered: &delivered}, noChatLogs{}, nil)
	result, err := useCase.Execute(ctx, GenerateAnswerCommand{
		Session: chatSessionFromRecord(sessionRecord), RequestUUID: "generate-answer-request", UserID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Answer.Answer != "durable answer" || delivered != "durable answer" {
		t.Fatalf("result=%#v delivered=%q", result.Answer, delivered)
	}
	request, err := q.GetChatRequest(ctx, sqlc_queries.GetChatRequestParams{
		Uuid: "generate-answer-request", ChatSessionUuid: sessionRecord.Uuid, UserID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Status != "completed" || request.AssistantUuid != "generate-answer-response" {
		t.Fatalf("workflow returned before durable completion: %#v", request)
	}
}
