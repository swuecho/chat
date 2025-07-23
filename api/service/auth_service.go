package service

import (
	"context"

	"github.com/swuecho/chat_backend/repository"
)

type authService struct {
	repos repository.CoreRepositoryManager
}

func NewAuthService(repos repository.CoreRepositoryManager) AuthService {
	return &authService{repos: repos}
}

func (s *authService) Login(ctx context.Context, username string, password string) (*LoginResult, error) {
	// TODO: Implement authentication logic
	return &LoginResult{
		AccessToken:  "placeholder-token",
		RefreshToken: "placeholder-refresh",
		ExpiresIn:    3600,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	// TODO: Implement token validation logic
	return &TokenClaims{
		UserID:   1,
		Username: "placeholder",
		IsAdmin:  false,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResult, error) {
	// TODO: Implement token refresh logic
	return &LoginResult{
		AccessToken:  "new-placeholder-token",
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	}, nil
}