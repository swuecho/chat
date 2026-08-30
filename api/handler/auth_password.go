package handler

import (
	"net/http"

	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/dto"
	"log/slog"
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
	if err := DecodeJSON(r, &req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("user"))
		return
	}

	tempPassword, err := auth.GenerateRandomPassword()
	if err != nil {
		slog.Error("Failed to generate temporary password", "error", err)
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to generate temporary password").WithDebugInfo(err.Error()))
		return
	}

	hashedPassword, err := auth.GeneratePasswordHash(tempPassword)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error()))
		return
	}

	if err := h.service.UpdateUserPassword(r.Context(), req.Email, hashedPassword); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to update password").WithDebugInfo(err.Error()))
		return
	}

	if err := sendPasswordResetEmail(user.Email, tempPassword); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to send password reset email").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// sendPasswordResetEmail sends a password reset email. Currently a no-op.
func sendPasswordResetEmail(email, tempPassword string) error {
	return nil
}

// ChangePasswordHandler updates the user's password.
func (h *AuthUserHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := DecodeJSON(r, &req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	// Verify the authenticated user owns the email being changed
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	user, err := h.service.GetAuthUserByID(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("user").WithDebugInfo(err.Error()))
		return
	}
	if user.Email != req.Email {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("Cannot change password for another user"))
		return
	}

	hashedPassword, err := auth.GeneratePasswordHash(req.NewPassword)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error()))
		return
	}

	if err := h.service.UpdateUserPassword(r.Context(), req.Email, string(hashedPassword)); err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update password"))
		return
	}

	w.WriteHeader(http.StatusOK)
}
