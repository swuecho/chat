package svc

import (
	"context"
	"testing"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestProviderModelRejectsMissingAndDisabledModels(t *testing.T) {
	ctx := context.Background()
	q := sqlc_queries.New(testDB)
	user, err := q.CreateAuthUser(ctx, sqlc_queries.CreateAuthUserParams{Email: "model-availability@test.com", Username: "availability", Password: "test"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteAuthUser(ctx, user.Email)
	model, err := q.CreateChatModel(ctx, sqlc_queries.CreateChatModelParams{Name: "availability-model", Label: "Availability", UserID: user.ID, ApiType: "openai"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteChatModel(ctx, sqlc_queries.DeleteChatModelParams{ID: model.ID, UserID: user.ID})
	if _, err := NewChatService(q, "", "").ProviderModel(ctx, model.Name); err != nil {
		t.Fatal(err)
	}
	if _, err := testDB.ExecContext(ctx, "UPDATE chat_model SET is_enable = false WHERE id = $1", model.ID); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{model.Name, "deleted-model"} {
		_, err := NewChatService(q, "", "").ProviderModel(ctx, name)
		appErr, ok := domain.AsError(err)
		if !ok || appErr.Kind != domain.KindNotFound || appErr.Resource != "Chat model" {
			t.Fatalf("model %s: unexpected error %v", name, err)
		}
	}
}
