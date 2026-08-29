package svc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// WorkspaceUnitOfWork contains only the operations used by workspace
// transactions. Simple query services continue to use SQLC directly.
type WorkspaceUnitOfWork interface {
	WorkspaceByUUID(context.Context, string) (Workspace, error)
	ClearDefaultWorkspaces(context.Context, int32) error
	SetDefaultWorkspace(context.Context, int32, string) (Workspace, error)
	CreateWorkspaceSession(context.Context, CreateWorkspaceSessionCommand) (ChatSession, error)
	EnsureSystemPrompt(context.Context, string, int32, string, string) error
	SetActiveSession(context.Context, int32, *int32, string) error
}

// SnapshotCopyUnitOfWork contains only the operations needed to copy a
// snapshot into a session.
type SnapshotCopyUnitOfWork interface {
	SnapshotByUUID(context.Context, string) (ChatSnapshot, error)
	PromptByUUID(context.Context, string) (ChatPrompt, error)
	InactiveSessionByUUID(context.Context, string) (ChatSession, error)
	CreateOrUpdateSession(context.Context, CreateOrUpdateChatSessionInput) (ChatSession, error)
	CreatePrompt(context.Context, CreateChatPromptInput) error
	CreateMessage(context.Context, CreateChatMessageInput) error
	SetActiveSession(context.Context, int32, *int32, string) error
}

type WorkspaceTransactionManager interface {
	WithinWorkspaceTransaction(context.Context, func(WorkspaceUnitOfWork) error) error
}

type SnapshotCopyTransactionManager interface {
	WithinSnapshotCopyTransaction(context.Context, func(SnapshotCopyUnitOfWork) error) error
}

type sqlcTransactionManager struct{ q *sqlc_queries.Queries }

func newSQLCTransactionManager(q *sqlc_queries.Queries) *sqlcTransactionManager {
	return &sqlcTransactionManager{q: q}
}

func (m *sqlcTransactionManager) WithinWorkspaceTransaction(ctx context.Context, fn func(WorkspaceUnitOfWork) error) error {
	return m.q.InTransaction(ctx, func(q *sqlc_queries.Queries) error {
		return fn(&sqlcUnitOfWork{q: q})
	})
}

func (m *sqlcTransactionManager) WithinSnapshotCopyTransaction(ctx context.Context, fn func(SnapshotCopyUnitOfWork) error) error {
	return m.q.InTransaction(ctx, func(q *sqlc_queries.Queries) error {
		return fn(&sqlcUnitOfWork{q: q})
	})
}

type sqlcUnitOfWork struct{ q *sqlc_queries.Queries }

func (u *sqlcUnitOfWork) WorkspaceByUUID(ctx context.Context, uuid string) (Workspace, error) {
	w, err := u.q.GetWorkspaceByUUID(ctx, uuid)
	return workspaceFromRecord(w), err
}

func (u *sqlcUnitOfWork) ClearDefaultWorkspaces(ctx context.Context, userID int32) error {
	return u.q.ClearDefaultWorkspacesByUserID(ctx, userID)
}

func (u *sqlcUnitOfWork) SetDefaultWorkspace(ctx context.Context, userID int32, uuid string) (Workspace, error) {
	w, err := u.q.SetDefaultWorkspaceForUser(ctx, sqlc_queries.SetDefaultWorkspaceForUserParams{Uuid: uuid, UserID: userID})
	return workspaceFromRecord(w), err
}

func (u *sqlcUnitOfWork) CreateWorkspaceSession(ctx context.Context, command CreateWorkspaceSessionCommand) (ChatSession, error) {
	workspaceID := sql.NullInt32{Int32: command.WorkspaceID, Valid: true}
	session, err := u.q.CreateChatSessionInWorkspace(ctx, sqlc_queries.CreateChatSessionInWorkspaceParams{
		UserID: command.UserID, Uuid: command.SessionUUID, Topic: command.Topic,
		Model: command.Model, MaxLength: 10, Active: true, WorkspaceID: workspaceID,
	})
	return chatSessionFromRecord(session), err
}

func (u *sqlcUnitOfWork) EnsureSystemPrompt(ctx context.Context, sessionUUID string, userID int32, text, promptUUID string) error {
	if _, err := u.q.GetOneChatPromptBySessionUUID(ctx, sessionUUID); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return eris.Wrap(err, "failed to check existing session prompt")
	}
	tokenCount := estimatePromptTokens(text)
	_, err := u.q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams{
		Uuid: promptUUID, ChatSessionUuid: sessionUUID, Role: "system", Content: text,
		TokenCount: tokenCount, UserID: userID, CreatedBy: userID, UpdatedBy: userID,
	})
	return eris.Wrap(err, "failed to create default system prompt")
}

func (u *sqlcUnitOfWork) SetActiveSession(ctx context.Context, userID int32, workspaceID *int32, sessionUUID string) error {
	var nullableWorkspaceID sql.NullInt32
	if workspaceID != nil {
		nullableWorkspaceID = sql.NullInt32{Int32: *workspaceID, Valid: true}
	}
	_, err := u.q.UpsertUserActiveSession(ctx, sqlc_queries.UpsertUserActiveSessionParams{
		UserID: userID, WorkspaceID: nullableWorkspaceID, ChatSessionUuid: sessionUUID,
	})
	return eris.Wrap(err, "failed to set active session")
}

func (u *sqlcUnitOfWork) SnapshotByUUID(ctx context.Context, uuid string) (ChatSnapshot, error) {
	r, err := u.q.ChatSnapshotByUUID(ctx, uuid)
	return chatSnapshotFromRecord(r), err
}

func (u *sqlcUnitOfWork) PromptByUUID(ctx context.Context, uuid string) (ChatPrompt, error) {
	r, err := u.q.GetChatPromptByUUID(ctx, uuid)
	return chatPromptFromRecord(r), err
}

func (u *sqlcUnitOfWork) InactiveSessionByUUID(ctx context.Context, uuid string) (ChatSession, error) {
	r, err := u.q.GetChatSessionByUUIDWithInActive(ctx, uuid)
	return chatSessionFromRecord(r), err
}

func (u *sqlcUnitOfWork) CreateOrUpdateSession(ctx context.Context, input CreateOrUpdateChatSessionInput) (ChatSession, error) {
	var workspaceID sql.NullInt32
	if input.WorkspaceID != nil {
		workspaceID = sql.NullInt32{Int32: *input.WorkspaceID, Valid: true}
	}
	r, err := u.q.CreateOrUpdateChatSessionByUUID(ctx, sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: input.UUID, UserID: input.UserID, Topic: input.Topic, MaxLength: input.MaxLength,
		Temperature: input.Temperature, Model: input.Model, MaxTokens: input.MaxTokens,
		TopP: input.TopP, N: input.N, Debug: input.Debug, SummarizeMode: input.SummarizeMode,
		WorkspaceID: workspaceID, ExploreMode: input.ExploreMode, ArtifactEnabled: input.ArtifactEnabled,
	})
	return chatSessionFromRecord(r), err
}

func (u *sqlcUnitOfWork) CreatePrompt(ctx context.Context, input CreateChatPromptInput) error {
	_, err := u.q.CreateChatPrompt(ctx, createChatPromptParams(input))
	return err
}

func (u *sqlcUnitOfWork) CreateMessage(ctx context.Context, input CreateChatMessageInput) error {
	_, err := u.q.CreateChatMessage(ctx, createChatMessageParams(input))
	return err
}
