package apicontract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/invopop/jsonschema"
	"github.com/swuecho/chat_backend/httpx"
)

// Info is the public metadata written into the OpenAPI document.
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// Registry stores contracts registered while the router is constructed.
type Registry struct {
	mu              sync.RWMutex
	info            Info
	operations      map[string]registeredOperation
	components      map[string]*jsonschema.Schema
	types           map[string]reflect.Type
	reflector       *jsonschema.Reflector
	pathPrefix      string
	securitySchemes map[string]any
}

type registeredOperation struct {
	operation      Operation
	requestSchema  any
	responseSchema any
}

// NewRegistry creates an empty OpenAPI 3.1 registry.
type Option func(*Registry)

// WithPathPrefix adds a router mount prefix to every documented path without
// changing the paths used to register handlers on a subrouter.
func WithPathPrefix(prefix string) Option {
	return func(r *Registry) { r.pathPrefix = strings.TrimSuffix(prefix, "/") }
}

// WithBearerAuth declares the bearerAuth security scheme used by protected
// operations and by Scalar's interactive API client.
func WithBearerAuth(description string) Option {
	return func(r *Registry) {
		r.securitySchemes["bearerAuth"] = map[string]any{
			"type": "http", "scheme": "bearer", "bearerFormat": "JWT", "description": description,
		}
	}
}

func NewRegistry(info Info, options ...Option) *Registry {
	registry := &Registry{
		info:            info,
		operations:      make(map[string]registeredOperation),
		components:      make(map[string]*jsonschema.Schema),
		types:           make(map[string]reflect.Type),
		securitySchemes: make(map[string]any),
		reflector: &jsonschema.Reflector{
			Anonymous:                  true,
			DoNotReference:             true,
			RequiredFromJSONSchemaTags: false,
			AllowAdditionalProperties:  false,
		},
	}
	for _, option := range options {
		option(registry)
	}
	return registry
}

func addOperation[Input, Output any](r *Registry, op Operation) error {
	if err := validateOperation(op); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := strings.ToUpper(op.Method) + " " + op.Path
	if _, exists := r.operations[key]; exists {
		return fmt.Errorf("duplicate api contract for %s", key)
	}
	for _, registered := range r.operations {
		if registered.operation.OperationID == op.OperationID {
			return fmt.Errorf("duplicate api contract operation ID %q", op.OperationID)
		}
	}

	requestSchema, err := r.schemaFor(reflect.TypeOf((*Input)(nil)).Elem())
	if err != nil {
		return fmt.Errorf("api contract %s request: %w", op.OperationID, err)
	}
	responseSchema, err := r.schemaFor(reflect.TypeOf((*Output)(nil)).Elem())
	if err != nil {
		return fmt.Errorf("api contract %s response: %w", op.OperationID, err)
	}
	r.operations[key] = registeredOperation{operation: op, requestSchema: requestSchema, responseSchema: responseSchema}
	return nil
}

func (r *Registry) schemaFor(t reflect.Type) (any, error) {
	if t == reflect.TypeOf(NoBody{}) {
		return nil, nil
	}
	if t == reflect.TypeOf(BinaryBody{}) {
		return map[string]any{"type": "string", "format": "binary"}, nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() == "" {
		return r.reflector.ReflectFromType(t), nil
	}
	name := t.Name()
	if existing, ok := r.types[name]; ok && existing != t {
		return nil, fmt.Errorf("schema name %q is shared by %s and %s", name, existing, t)
	}
	if _, ok := r.components[name]; !ok {
		r.components[name] = r.reflector.ReflectFromType(t)
		r.types[name] = t
	}
	return map[string]any{"$ref": "#/components/schemas/" + name}, nil
}

// Handler serves the generated OpenAPI document.
func (r *Registry) Handler() http.HandlerFunc {
	return httpx.Adapt(func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, http.StatusOK, r.Document())
	})
}

// Document returns a deterministic OpenAPI 3.1 document snapshot.
func (r *Registry) Document() Document {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paths := make(map[string]map[string]operationObject)
	keys := make([]string, 0, len(r.operations))
	for key := range r.operations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		registered := r.operations[key]
		op := registered.operation
		parameters := make([]parameterObject, 0, len(op.Parameters))
		for _, parameter := range op.Parameters {
			parameters = append(parameters, parameterObject{Name: parameter.Name, In: parameter.In,
				Description: parameter.Description, Required: parameter.Required, Schema: parameter.Schema})
		}
		operation := operationObject{OperationID: op.OperationID, Summary: op.Summary,
			Description: op.Description, Tags: op.Tags, Parameters: parameters, Security: op.Security,
			Responses: map[string]responseObject{fmt.Sprint(op.SuccessStatus): {
				Description: http.StatusText(op.SuccessStatus),
			}}}
		if registered.requestSchema != nil {
			operation.RequestBody = &requestBodyObject{Required: true, Content: mediaContent(op.RequestContentType, registered.requestSchema)}
		}
		response := operation.Responses[fmt.Sprint(op.SuccessStatus)]
		if registered.responseSchema != nil {
			response.Content = mediaContent(op.ResponseContentType, registered.responseSchema)
		}
		operation.Responses[fmt.Sprint(op.SuccessStatus)] = response
		operation.Responses["default"] = responseObject{Description: "API error", Content: jsonContent(map[string]any{"$ref": "#/components/schemas/APIError"})}

		path := r.pathPrefix + op.Path
		if paths[path] == nil {
			paths[path] = make(map[string]operationObject)
		}
		paths[path][strings.ToLower(op.Method)] = operation
	}

	components := make(map[string]any, len(r.components)+1)
	components["APIError"] = map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{
			"code": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"},
			"detail": map[string]any{"type": "string"},
		}, "required": []string{"code", "message"},
	}
	for name, schema := range r.components {
		components[name] = schema
	}
	return Document{OpenAPI: "3.1.0", Info: r.info, Paths: paths, Components: componentsObject{
		Schemas: components, SecuritySchemes: r.securitySchemes,
	}}
}

// MarshalJSON is a convenience for generators and golden tests.
func (r *Registry) MarshalJSON() ([]byte, error) { return json.MarshalIndent(r.Document(), "", "  ") }

// Document is the serializable OpenAPI root object.
type Document struct {
	OpenAPI    string                                `json:"openapi"`
	Info       Info                                  `json:"info"`
	Paths      map[string]map[string]operationObject `json:"paths"`
	Components componentsObject                      `json:"components"`
}

type componentsObject struct {
	Schemas         map[string]any `json:"schemas"`
	SecuritySchemes map[string]any `json:"securitySchemes,omitempty"`
}
type operationObject struct {
	OperationID string                    `json:"operationId"`
	Summary     string                    `json:"summary,omitempty"`
	Description string                    `json:"description,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Parameters  []parameterObject         `json:"parameters,omitempty"`
	RequestBody *requestBodyObject        `json:"requestBody,omitempty"`
	Responses   map[string]responseObject `json:"responses"`
	Security    []map[string][]string     `json:"security,omitempty"`
}
type parameterObject struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      map[string]any `json:"schema"`
}
type requestBodyObject struct {
	Required bool                       `json:"required"`
	Content  map[string]mediaTypeObject `json:"content"`
}
type responseObject struct {
	Description string                     `json:"description"`
	Content     map[string]mediaTypeObject `json:"content,omitempty"`
}
type mediaTypeObject struct {
	Schema any `json:"schema"`
}

func jsonContent(schema any) map[string]mediaTypeObject {
	return map[string]mediaTypeObject{"application/json": {Schema: schema}}
}

func mediaContent(contentType string, schema any) map[string]mediaTypeObject {
	if contentType == "" {
		contentType = "application/json"
	}
	return map[string]mediaTypeObject{contentType: {Schema: schema}}
}
