package handler

import (
	"encoding/json"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
)

func TestWorkspaceHandlerRegistersTypedContracts(t *testing.T) {
	registry := apicontract.NewRegistry(apicontract.Info{Title: "Test", Version: "1"}, apicontract.WithPathPrefix("/api"))
	router := mux.NewRouter()
	NewChatWorkspaceHandler(nil).Register(router, registry)

	document, err := registry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	paths := decoded["paths"].(map[string]any)
	for _, path := range []string{
		"/api/workspaces",
		"/api/workspaces/default",
		"/api/workspaces/auto-migrate",
		"/api/workspaces/{uuid}",
		"/api/workspaces/{uuid}/reorder",
		"/api/workspaces/{uuid}/set-default",
		"/api/workspaces/{uuid}/sessions",
	} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("workspace path %s missing from OpenAPI document", path)
		}
	}
}
