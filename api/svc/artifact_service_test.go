package svc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestArtifactServiceRejectsInvalidUpdates(t *testing.T) {
	service := NewArtifactService(nil)
	for _, test := range []struct {
		name    string
		command UpdateArtifactCommand
	}{
		{name: "blank title", command: UpdateArtifactCommand{Title: "  "}},
		{name: "long title", command: UpdateArtifactCommand{Title: strings.Repeat("a", maxArtifactTitleBytes+1)}},
		{name: "large content", command: UpdateArtifactCommand{Title: "Valid", Content: strings.Repeat("a", maxArtifactContentBytes+1)}},
		{name: "long language", command: UpdateArtifactCommand{Title: "Valid", Language: strings.Repeat("a", 65)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := service.Update(context.Background(), test.command); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestArtifactServiceCatalogAndMutations(t *testing.T) {
	ctx := context.Background()
	q := sqlc_queries.New(testDB)
	sessionUUID := util.NewUUID()
	messageUUID := util.NewUUID()
	artifactUUID := util.NewUUID()
	session, err := q.CreateChatSession(ctx, sqlc_queries.CreateChatSessionParams{UserID: 1, Topic: "Artifact catalog test", MaxLength: 10, Uuid: sessionUUID, Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.DeleteChatSession(ctx, session.ID)
	_, err = q.CreateChatMessage(ctx, sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid: sessionUUID, Uuid: messageUUID, Role: "assistant", Content: "artifact test", Model: "test-model",
		UserID: 1, CreatedBy: 1, UpdatedBy: 1, Raw: json.RawMessage(`{}`), SuggestedQuestions: json.RawMessage(`[]`),
		Artifacts: json.RawMessage(`[{"uuid":"` + artifactUUID + `","type":"html","title":"` + artifactUUID + `","content":"<p>one</p>","language":"html"}]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.ExecContext(ctx, "UPDATE chat_message SET is_deleted = true WHERE uuid = $1", messageUUID)

	service := NewArtifactService(q)
	page, err := service.List(ctx, ArtifactPageQuery{UserID: 1, Search: artifactUUID, Limit: 20})
	if err != nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("List() = %#v, %v", page, err)
	}
	if err := service.Update(ctx, UpdateArtifactCommand{UUID: artifactUUID, Title: "Forbidden", Content: "changed", Language: "html", UserID: 2}); err == nil {
		t.Fatal("another user updated the artifact")
	}
	if _, err := service.Duplicate(ctx, artifactUUID, 2); err == nil {
		t.Fatal("another user duplicated the artifact")
	}
	if err := service.Delete(ctx, artifactUUID, 2); err == nil {
		t.Fatal("another user deleted the artifact")
	}

	if err := service.Update(ctx, UpdateArtifactCommand{UUID: artifactUUID, Title: "  Updated  ", Content: "<p>two</p>", Language: " HTML ", UserID: 1}); err != nil {
		t.Fatal(err)
	}
	duplicateUUID, err := service.Duplicate(ctx, artifactUUID, 1)
	if err != nil || duplicateUUID == "" {
		t.Fatalf("Duplicate() uuid = %q, err = %v", duplicateUUID, err)
	}
	if err := service.Delete(ctx, artifactUUID, 1); err != nil {
		t.Fatal(err)
	}

	page, err = service.List(ctx, ArtifactPageQuery{UserID: 1, SessionUUID: sessionUUID, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	foundDuplicate := false
	for _, item := range page.Items {
		if item.UUID == duplicateUUID && item.Title == "Updated (Copy)" && item.Content == "<p>two</p>" {
			foundDuplicate = true
		}
		if item.UUID == artifactUUID {
			t.Fatal("deleted artifact remains in catalog")
		}
	}
	if !foundDuplicate {
		t.Fatal("duplicated artifact missing from catalog")
	}
}
