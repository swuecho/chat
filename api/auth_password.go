// Package main — Password reset and change handlers.
package main

import (
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// --- Request types ---

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

type ChangePasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

// --- Handlers ---

// ResetPasswordHandler generates a temporary password and sends it via email.
func (h *AuthUserHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("user"))
		return
	}

	tempPassword, err := auth.GenerateRandomPassword()
	if err != nil {
		log.WithError(err).Error("Failed to generate temporary password")
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to generate temporary password").WithDebugInfo(err.Error()))
		return
	}

	hashedPassword, err := auth.GeneratePasswordHash(tempPassword)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error()))
		return
	}

	if err := h.service.UpdateUserPassword(context.Background(), sqlc_queries.UpdateUserPasswordParams{
		Email: req.Email, Password: hashedPassword,
	}); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to update password").WithDebugInfo(err.Error()))
		return
	}

	if err := SendPasswordResetEmail(user.Email, tempPassword); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to send password reset email").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SendPasswordResetEmail sends a password reset email. Currently a no-op.
func SendPasswordResetEmail(email, tempPassword string) error {
	return nil
}

// ChangePasswordHandler updates the user's password.
func (h *AuthUserHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	hashedPassword, err := auth.GeneratePasswordHash(req.NewPassword)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error()))
		return
	}

	if err := h.service.UpdateUserPassword(context.Background(), sqlc_queries.UpdateUserPasswordParams{
		Email: req.Email, Password: string(hashedPassword),
	}); err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update password"))
		return
	}

	w.WriteHeader(http.StatusOK)
}
