package svc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelPrivilegeService struct{ q *sqlc_queries.Queries }

func NewChatModelPrivilegeService(q *sqlc_queries.Queries) *ChatModelPrivilegeService {
	return &ChatModelPrivilegeService{q: q}
}

type ChatModelPrivilege struct {
	ID            int32  `json:"id"`
	FullName      string `json:"fullName"`
	UserEmail     string `json:"userEmail"`
	ChatModelName string `json:"chatModelName"`
	RateLimit     int32  `json:"rateLimit"`
}

func (s *ChatModelPrivilegeService) List(ctx context.Context) ([]ChatModelPrivilege, error) {
	rows, err := s.q.ListUserChatModelPrivilegesRateLimit(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ChatModelPrivilege, len(rows))
	for i, row := range rows {
		result[i] = ChatModelPrivilege{ID: row.ID, FullName: row.FullName, UserEmail: row.UserEmail, ChatModelName: row.ChatModelName, RateLimit: row.RateLimit}
	}
	return result, nil
}

func (s *ChatModelPrivilegeService) Create(ctx context.Context, input ChatModelPrivilege) (ChatModelPrivilege, error) {
	user, err := s.q.GetAuthUserByEmail(ctx, input.UserEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChatModelPrivilege{}, domain.NotFound("user", err)
		}
		return ChatModelPrivilege{}, err
	}
	model, err := s.q.ChatModelByName(ctx, input.ChatModelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChatModelPrivilege{}, domain.NotFound("Chat model", err)
		}
		return ChatModelPrivilege{}, err
	}
	created, err := s.q.CreateUserChatModelPrivilege(ctx, sqlc_queries.CreateUserChatModelPrivilegeParams{
		UserID: user.ID, ChatModelID: model.ID, RateLimit: input.RateLimit, CreatedBy: user.ID, UpdatedBy: user.ID,
	})
	if err != nil {
		return ChatModelPrivilege{}, err
	}
	return ChatModelPrivilege{ID: created.ID, UserEmail: user.Email, ChatModelName: model.Name, RateLimit: created.RateLimit}, nil
}

func (s *ChatModelPrivilegeService) Update(ctx context.Context, id, rateLimit, updatedBy int32, userEmail, modelName string) (ChatModelPrivilege, error) {
	updated, err := s.q.UpdateUserChatModelPrivilege(ctx, sqlc_queries.UpdateUserChatModelPrivilegeParams{ID: id, RateLimit: rateLimit, UpdatedBy: updatedBy})
	if err != nil {
		return ChatModelPrivilege{}, err
	}
	return ChatModelPrivilege{ID: updated.ID, UserEmail: userEmail, ChatModelName: modelName, RateLimit: updated.RateLimit}, nil
}

func (s *ChatModelPrivilegeService) Delete(ctx context.Context, id int32) error {
	return s.q.DeleteUserChatModelPrivilege(ctx, id)
}

func (s *ChatModelPrivilegeService) ListByUserID(ctx context.Context, userID int32) ([]sqlc_queries.UserChatModelPrivilege, error) {
	return s.q.ListUserChatModelPrivilegesByUserID(ctx, userID)
}
