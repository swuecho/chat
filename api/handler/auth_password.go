package handler

import (
	"net/http"

	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/dto"
)

// --- Request types ---

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

func (r *ResetPasswordRequest) Validate() error {
	if r.Email == "" {
		return dto.ErrValidationInvalidInput("email is required")
	}
	return nil
}

type ChangePasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

func (r *ChangePasswordRequest) Validate() error {
	if r.Email == "" || r.NewPassword == "" {
		return dto.ErrValidationInvalidInput("email and new_password are required")
	}
	return nil
}

// --- Handlers ---

// ResetPasswordHandler generates a temporary password and sends it via email.
func (h *AuthUserHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) error {
	var req ResetPasswordRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		return dto.ErrResourceNotFound("user")
	}

	tempPassword, err := auth.GenerateRandomPassword()
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to generate temporary password").WithDebugInfo(err.Error())
	}

	hashedPassword, err := auth.GeneratePasswordHash(tempPassword)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error())
	}

	if err := h.service.UpdateUserPassword(r.Context(), req.Email, hashedPassword); err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to update password").WithDebugInfo(err.Error())
	}

	if err := sendPasswordResetEmail(user.Email, tempPassword); err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to send password reset email").WithDebugInfo(err.Error())
	}
	return respondStatus(w, http.StatusOK)
}

// sendPasswordResetEmail sends a password reset email. Currently a no-op.
func sendPasswordResetEmail(email, tempPassword string) error {
	return nil
}

// ChangePasswordHandler updates the user's password.
func (h *AuthUserHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) error {
	var req ChangePasswordRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	// Verify the authenticated user owns the email being changed
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	user, err := h.service.GetAuthUserByID(r.Context(), userID)
	if err != nil {
		return dto.ErrResourceNotFound("user").WithDebugInfo(err.Error())
	}
	if user.Email != req.Email {
		return dto.ErrAuthAccessDenied.WithMessage("Cannot change password for another user")
	}

	hashedPassword, err := auth.GeneratePasswordHash(req.NewPassword)
	if err != nil {
		return dto.ErrInternalUnexpected.WithMessage("Failed to hash password").WithDebugInfo(err.Error())
	}

	if err := h.service.UpdateUserPassword(r.Context(), req.Email, string(hashedPassword)); err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update password")
	}
	return respondStatus(w, http.StatusOK)
}
