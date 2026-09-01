package handler

import (
	"encoding/json"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apiopenapi"
)

func TestAPIKeyOpenAPIUsesJSONWireTypes(t *testing.T) {
	registry := apiopenapi.NewRegistry()
	NewAPIKeyHandler(nil).Register(mux.NewRouter(), registry)
	document, err := registry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var spec map[string]any
	if err := json.Unmarshal(document, &spec); err != nil {
		t.Fatal(err)
	}
	components := spec["components"].(map[string]any)["schemas"].(map[string]any)
	detail := components["gatewayRequestDetailHTTPResponse"].(map[string]any)["properties"].(map[string]any)

	requestUUID := detail["requestUuid"].(map[string]any)
	if requestUUID["type"] != "string" || requestUUID["format"] != "uuid" {
		t.Fatalf("requestUuid schema = %#v", requestUUID)
	}
	completedAt := detail["completedAt"].(map[string]any)
	if _, ok := completedAt["oneOf"]; !ok {
		t.Fatalf("completedAt is not nullable: %#v", completedAt)
	}

	createdKey := components["CreatedAPIKey"].(map[string]any)["properties"].(map[string]any)
	if _, ok := createdKey["expiresAt"].(map[string]any)["oneOf"]; !ok {
		t.Fatalf("expiresAt is not nullable: %#v", createdKey["expiresAt"])
	}
	if _, ok := createdKey["lastUsedAt"].(map[string]any)["oneOf"]; !ok {
		t.Fatalf("lastUsedAt is not nullable: %#v", createdKey["lastUsedAt"])
	}
}
