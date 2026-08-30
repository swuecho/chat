package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestValidateUUIDRouteParams(t *testing.T) {
	router := mux.NewRouter()
	router.Use(ValidateUUIDRouteParams)
	router.HandleFunc("/sessions/{sessionUUID}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	valid := httptest.NewRecorder()
	router.ServeHTTP(valid, httptest.NewRequest(http.MethodGet, "/sessions/01990a45-8a36-7e51-bf7c-a8df8d6b8e91", nil))
	if valid.Code != http.StatusNoContent {
		t.Fatalf("valid UUID status = %d", valid.Code)
	}

	invalid := httptest.NewRecorder()
	router.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, "/sessions/not-a-uuid", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid UUID status = %d", invalid.Code)
	}
}
