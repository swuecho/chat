package handler

import (
	"encoding/json"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
)

func TestArtifactHandlerRegistersTypedContracts(t *testing.T) {
	registry := apicontract.NewRegistry(apicontract.Info{Title: "Test", Version: "1"}, apicontract.WithPathPrefix("/api"))
	NewArtifactHandler(nil).Register(mux.NewRouter(), registry)
	document, err := registry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	paths := decoded["paths"].(map[string]any)
	for _, path := range []string{"/api/artifacts", "/api/artifacts/{uuid}", "/api/artifacts/{uuid}/duplicate"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("artifact path %s missing from OpenAPI document", path)
		}
	}
}
