// Package svc provides business logic services for the chat application.
package svc

// Cfg holds configuration set by main during initialization.
var Cfg struct {
	OpenAIKey    string
	OpenAIProxy  string
	JWTSecret    string
	DefaultLimit int32
}
