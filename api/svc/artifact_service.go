package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

const (
	maxArtifactTitleBytes   = 200
	maxArtifactContentBytes = 1024 * 1024
)

type ArtifactRecord struct {
	domain.Artifact
	MessageUUID, SessionUUID, SessionTitle string
	CreatedAt, UpdatedAt                   time.Time
}

type ArtifactPageQuery struct {
	UserID, Limit, Offset               int32
	Search, Type, Language, SessionUUID string
}

type ArtifactPage struct {
	Items         []ArtifactRecord
	Total         int64
	Limit, Offset int32
}

type UpdateArtifactCommand struct {
	UUID, Title, Content, Language string
	UserID                         int32
}

type ArtifactService struct{ q *sqlc_queries.Queries }

func NewArtifactService(q *sqlc_queries.Queries) *ArtifactService { return &ArtifactService{q: q} }

func (s *ArtifactService) List(ctx context.Context, query ArtifactPageQuery) (ArtifactPage, error) {
	params := sqlc_queries.ListArtifactsParams{UserID: query.UserID, Search: query.Search, ArtifactType: query.Type,
		Language: query.Language, SessionUuid: query.SessionUUID, PageLimit: query.Limit, PageOffset: query.Offset}
	rows, err := s.q.ListArtifacts(ctx, params)
	if err != nil {
		return ArtifactPage{}, err
	}
	count, err := s.q.CountArtifacts(ctx, sqlc_queries.CountArtifactsParams{UserID: query.UserID, Search: query.Search,
		ArtifactType: query.Type, Language: query.Language, SessionUuid: query.SessionUUID})
	if err != nil {
		return ArtifactPage{}, err
	}
	items := make([]ArtifactRecord, 0, len(rows))
	for _, row := range rows {
		var artifact domain.Artifact
		if err := json.Unmarshal([]byte(row.ArtifactJson), &artifact); err != nil {
			return ArtifactPage{}, err
		}
		items = append(items, ArtifactRecord{Artifact: artifact, MessageUUID: row.MessageUuid, SessionUUID: row.SessionUuid,
			SessionTitle: row.SessionTitle, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return ArtifactPage{Items: items, Total: count, Limit: query.Limit, Offset: query.Offset}, nil
}

func (s *ArtifactService) Update(ctx context.Context, command UpdateArtifactCommand) error {
	command.Title = strings.TrimSpace(command.Title)
	command.Language = strings.ToLower(strings.TrimSpace(command.Language))
	if command.Title == "" {
		return domain.Invalid("artifact title is required")
	}
	if len(command.Title) > maxArtifactTitleBytes {
		return domain.Invalid("artifact title is too long")
	}
	if len(command.Content) > maxArtifactContentBytes {
		return domain.Invalid("artifact content exceeds the 1 MB limit")
	}
	if len(command.Language) > 64 {
		return domain.Invalid("artifact language is too long")
	}
	_, err := s.q.UpdateArtifact(ctx, sqlc_queries.UpdateArtifactParams{ArtifactUuid: command.UUID, Title: command.Title,
		Content: command.Content, Language: command.Language, UserID: command.UserID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotFound("Artifact", err)
	}
	return err
}

func (s *ArtifactService) Delete(ctx context.Context, uuid string, userID int32) error {
	_, err := s.q.DeleteArtifact(ctx, sqlc_queries.DeleteArtifactParams{ArtifactUuid: uuid, UserID: userID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NotFound("Artifact", err)
	}
	return err
}

func (s *ArtifactService) Duplicate(ctx context.Context, uuid string, userID int32) (string, error) {
	newUUID := util.NewUUID()
	_, err := s.q.DuplicateArtifact(ctx, sqlc_queries.DuplicateArtifactParams{ArtifactUuid: uuid, NewUuid: newUUID, UserID: userID})
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.NotFound("Artifact", err)
	}
	return newUUID, err
}
