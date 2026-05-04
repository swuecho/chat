// Package app provides the central application context holding all shared dependencies.
// It replaces package-level global state with dependency injection.
package app

import (
	"database/sql"

	"golang.org/x/time/rate"
)

// App holds all shared application dependencies.
// Pass this to constructors instead of relying on global variables.
type App struct {
	DB          *sql.DB
	Config      Config
	RateLimiter *rate.Limiter
}

// Config holds runtime configuration values needed across packages.
type Config struct {
	OpenAIKey    string
	OpenAIProxy  string
	JWTSecret     string
	JWTAudience   string
	DefaultLimit  int32
}
