package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"github.com/swuecho/chat_backend/svc"
)

// Snapshots written before the application-model migration encoded workspaceId
// as sql.NullInt32. Keep those saved snapshots readable alongside current ones.
func decodeSnapshotSession(data json.RawMessage) (svc.ChatSession, error) {
	var stored struct {
		chatSessionHTTPResponse
		Workspace json.RawMessage `json:"workspaceId"`
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		return svc.ChatSession{}, err
	}
	var workspaceID *int32
	workspace := bytes.TrimSpace(stored.Workspace)
	if len(workspace) > 0 && workspace[0] == '{' {
		var legacy sql.NullInt32
		if err := json.Unmarshal(workspace, &legacy); err != nil {
			return svc.ChatSession{}, err
		}
		if legacy.Valid {
			workspaceID = &legacy.Int32
		}
	} else if len(workspace) > 0 {
		if err := json.Unmarshal(workspace, &workspaceID); err != nil {
			return svc.ChatSession{}, err
		}
	}
	s := stored.chatSessionHTTPResponse
	return svc.ChatSession{ID: s.ID, UserID: s.UserID, UUID: s.UUID,
		Topic: s.Topic, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
		Active: s.Active, Model: s.Model, MaxLength: s.MaxLength,
		Temperature: s.Temperature, TopP: s.TopP, MaxTokens: s.MaxTokens,
		N: s.N, SummarizeMode: s.SummarizeMode, WorkspaceID: workspaceID,
		ArtifactEnabled: s.ArtifactEnabled, ExploreMode: s.ExploreMode}, nil
}
