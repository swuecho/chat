// Package apiopenapi owns application-wide OpenAPI configuration shared by the
// running server and the build-time generator.
package apiopenapi

import "github.com/swuecho/chat_backend/apicontract"

func NewRegistry() *apicontract.Registry {
	return apicontract.NewRegistry(apicontract.Info{
		Title:       "Chat API",
		Description: "Multi-LLM chat service API",
		Version:     "1.0.0",
	}, apicontract.WithPathPrefix("/api"), apicontract.WithBearerAuth("Browser JWT access token"))
}
