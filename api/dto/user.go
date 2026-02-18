package dto

// LoginParams represents login credentials
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenRequest represents a token request
type TokenRequest struct {
	Token string `json:"token"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Email string `json:"email"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// UserStat represents user statistics
type UserStat struct {
	SessionCount    int64 `json:"sessionCount"`
	MessageCount    int64 `json:"messageCount"`
	SnapshotCount   int64 `json:"snapshotCount"`
	WorkspaceCount  int64 `json:"workspaceCount"`
	PromptCount     int64 `json:"promptCount"`
	FileCount       int64 `json:"fileCount"`
	TotalTokenCount int64 `json:"totalTokenCount"`
}

// RateLimitRequest represents a rate limit configuration request
type RateLimitRequest struct {
	RateLimit int `json:"rateLimit"`
}
