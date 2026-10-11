package handler

import (
	"encoding/json"
	"testing"

	"github.com/swuecho/chat_backend/svc"
)

func TestDecodeSnapshotSession(t *testing.T) {
	for _, tt := range []struct {
		name      string
		workspace string
		wantID    bool
	}{
		{"legacy valid", `{"Int32":42,"Valid":true}`, true},
		{"legacy null", `{"Int32":42,"Valid":false}`, false},
		{"current", `42`, true},
		{"null", `null`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			data := json.RawMessage(`{"uuid":"session","model":"test","temperature":0.7,"maxTokens":123,"workspaceId":` + tt.workspace + `}`)
			s, err := decodeSnapshotSession(data)
			if err != nil {
				t.Fatal(err)
			}
			if s.UUID != "session" || s.Model != "test" || s.Temperature != 0.7 || s.MaxTokens != 123 {
				t.Fatalf("lost session settings: %+v", s)
			}
			if (s.WorkspaceID != nil) != tt.wantID {
				t.Fatalf("unexpected workspace: %v", s.WorkspaceID)
			}
			if tt.wantID && *s.WorkspaceID != 42 {
				t.Fatalf("workspace = %d", *s.WorkspaceID)
			}
		})
	}
	for _, data := range []string{`{}`, `{"model":"test"}`} {
		if _, err := decodeSnapshotSession(json.RawMessage(data)); err != nil {
			t.Fatal(err)
		}
	}
	for _, data := range []string{`{`, `{"workspaceId":"invalid"}`, `{"workspaceId":{"Int32":"invalid","Valid":true}}`} {
		if _, err := decodeSnapshotSession(json.RawMessage(data)); err == nil {
			t.Fatalf("accepted invalid snapshot: %s", data)
		}
	}
}

func TestDecodeSnapshotSessionRoundTrip(t *testing.T) {
	id := int32(42)
	want := svc.ChatSession{UUID: "session", Model: "test", WorkspaceID: &id, TopP: 0.9, SummarizeMode: true, ArtifactEnabled: true, ExploreMode: true}
	got, err := decodeSnapshotSession(*want.ToRawMessage())
	if err != nil {
		t.Fatal(err)
	}
	if got.UUID != want.UUID || got.Model != want.Model || got.TopP != want.TopP || !got.SummarizeMode || !got.ArtifactEnabled || !got.ExploreMode || got.WorkspaceID == nil || *got.WorkspaceID != id {
		t.Fatalf("round trip lost settings: %+v", got)
	}
}
