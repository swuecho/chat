// Package apicontract couples HTTP route registration with the request and
// response types used to produce the service's OpenAPI document.
package apicontract

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/httpx"
)

// NoBody marks an operation that has no JSON request body or response body.
type NoBody struct{}

// Operation describes the HTTP-specific metadata that cannot be inferred from
// Go request and response types.
type Operation struct {
	Method        string
	Path          string
	OperationID   string
	Summary       string
	Description   string
	Tags          []string
	SuccessStatus int
	Parameters    []Parameter
	Security      []map[string][]string
}

// Parameter documents a path, query, or header parameter.
type Parameter struct {
	Name        string
	In          string
	Description string
	Required    bool
	Schema      map[string]any
}

// UUIDPathParameter describes a required UUID path segment.
func UUIDPathParameter(name string) Parameter {
	return Parameter{Name: name, In: "path", Required: true, Schema: map[string]any{"type": "string", "format": "uuid"}}
}

// BearerAuth requires the standard bearer authentication scheme configured on
// the registry.
func BearerAuth() []map[string][]string {
	return []map[string][]string{{"bearerAuth": {}}}
}

// Handler is a typed JSON transport handler. HTTP authentication and route
// parameter extraction remain available through the request.
type Handler[Input, Output any] func(*http.Request, Input) (Output, error)

// RegisterJSON installs the runtime handler and records the same input/output
// types in the OpenAPI registry. Invalid or duplicate contracts panic during
// startup, like an invalid Gorilla Mux route declaration.
func RegisterJSON[Input, Output any](router *mux.Router, registry *Registry, op Operation, handler Handler[Input, Output]) {
	if err := addOperation[Input, Output](registry, op); err != nil {
		panic(err)
	}

	router.HandleFunc(op.Path, httpx.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		var input Input
		if !isNoBody[Input]() {
			if err := httpx.DecodeJSON(r, &input); err != nil {
				return err
			}
		}

		output, err := handler(r, input)
		if err != nil {
			return err
		}
		if isNoBody[Output]() {
			return httpx.Status(w, op.SuccessStatus)
		}
		return httpx.JSON(w, op.SuccessStatus, output)
	})).Methods(op.Method)
}

func validateOperation(op Operation) error {
	if op.Method == "" || op.Path == "" || op.OperationID == "" {
		return fmt.Errorf("api contract requires method, path, and operation ID")
	}
	if !strings.HasPrefix(op.Path, "/") {
		return fmt.Errorf("api contract path %q must begin with /", op.Path)
	}
	if op.SuccessStatus < 200 || op.SuccessStatus > 299 {
		return fmt.Errorf("api contract %s success status must be 2xx", op.OperationID)
	}
	for _, parameter := range op.Parameters {
		if parameter.Name == "" || (parameter.In != "path" && parameter.In != "query" && parameter.In != "header") {
			return fmt.Errorf("api contract %s has invalid parameter", op.OperationID)
		}
		if parameter.In == "path" && !parameter.Required {
			return fmt.Errorf("api contract %s path parameter %s must be required", op.OperationID, parameter.Name)
		}
	}
	return nil
}

func isNoBody[T any]() bool {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return t == reflect.TypeOf(NoBody{})
}
