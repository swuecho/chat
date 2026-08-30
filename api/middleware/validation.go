package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/validation"
)

// ValidateUUIDRouteParams rejects malformed UUID route parameters before the
// request reaches a handler or application service. UUID parameters follow the
// project's existing naming convention (uuid, sessionUUID, bot_uuid, etc.).
func ValidateUUIDRouteParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, value := range mux.Vars(r) {
			if !strings.Contains(strings.ToLower(name), "uuid") {
				continue
			}
			if err := validation.UUID(name, value, true); err != nil {
				dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput(err.Error()))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
