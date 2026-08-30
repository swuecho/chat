package apicontract_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
)

type createWidgetRequest struct {
	Name string `json:"name" jsonschema:"required,minLength=1,maxLength=20"`
}

func (r *createWidgetRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

type widgetResponse struct {
	ID   string `json:"id" jsonschema:"format=uuid"`
	Name string `json:"name"`
}

func TestRegisterJSONUsesTypedRuntimePipelineAndDocument(t *testing.T) {
	registry := apicontract.NewRegistry(apicontract.Info{Title: "Test", Version: "1"}, apicontract.WithPathPrefix("/api"))
	router := mux.NewRouter()
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodPost, Path: "/widgets", OperationID: "createWidget",
		SuccessStatus: http.StatusCreated,
	}, func(_ *http.Request, input createWidgetRequest) (widgetResponse, error) {
		return widgetResponse{ID: "00000000-0000-0000-0000-000000000001", Name: input.Name}, nil
	})

	request := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"name":"desk"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	document, err := registry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	paths := decoded["paths"].(map[string]any)
	operation := paths["/api/widgets"].(map[string]any)["post"].(map[string]any)
	if operation["operationId"] != "createWidget" {
		t.Fatalf("operationId = %v", operation["operationId"])
	}
	components := decoded["components"].(map[string]any)["schemas"].(map[string]any)
	requestSchema := components["createWidgetRequest"].(map[string]any)
	properties := requestSchema["properties"].(map[string]any)
	if properties["name"].(map[string]any)["maxLength"] != float64(20) {
		t.Fatalf("request constraint missing: %s", document)
	}
	if _, ok := components["widgetResponse"]; !ok {
		t.Fatalf("response component missing: %s", document)
	}
}

func TestRegisterJSONRejectsUnknownFieldsBeforeHandler(t *testing.T) {
	registry := apicontract.NewRegistry(apicontract.Info{Title: "Test", Version: "1"})
	router := mux.NewRouter()
	called := false
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodPost, Path: "/widgets", OperationID: "createWidget", SuccessStatus: http.StatusCreated,
	}, func(_ *http.Request, input createWidgetRequest) (widgetResponse, error) {
		called = true
		return widgetResponse{Name: input.Name}, nil
	})

	request := httptest.NewRequest(http.MethodPost, "/widgets", strings.NewReader(`{"name":"desk","unknown":true}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if called {
		t.Fatal("handler called for invalid request")
	}
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestRegisterJSONSupportsBodylessOperations(t *testing.T) {
	registry := apicontract.NewRegistry(apicontract.Info{Title: "Test", Version: "1"})
	router := mux.NewRouter()
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodDelete, Path: "/widgets/{id}", OperationID: "deleteWidget", SuccessStatus: http.StatusNoContent,
	}, func(_ *http.Request, _ apicontract.NoBody) (apicontract.NoBody, error) {
		return apicontract.NoBody{}, nil
	})

	request := httptest.NewRequest(http.MethodDelete, "/widgets/1", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	document, err := registry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(document), "requestBody") {
		t.Fatalf("bodyless operation documented a request body: %s", document)
	}
}

func TestScalarHandlerLoadsGeneratedDocument(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	response := httptest.NewRecorder()
	apicontract.ScalarHandler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	body := response.Body.String()
	for _, expected := range []string{"@scalar/api-reference", "url: '/api/openapi.json'"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("Scalar page does not contain %q", expected)
		}
	}
}
