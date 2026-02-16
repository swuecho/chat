// Package main provides constants used throughout the chat application.
// This file contains all magic numbers, timeouts, and configuration values
// to improve code maintainability and avoid scattered constants.
package main

import "time"

// API and request constants
const (
	// Timeout settings
	DefaultRequestTimeout = 5 * time.Minute

	// Loop limits and safety guards
	MaxStreamingLoopIterations = 10000

	// Content buffering and flushing
	SmallAnswerThreshold    = 200
	FlushCharacterThreshold = 500
	TestPrefixLength        = 16

	// Pagination
	DefaultPageSize = 200
	MaxHistoryItems = 10000

	// Rate limiting
	DefaultPageLimit = 30

	// Test constants
	TestDemoPrefix = "test_demo_bestqa"

	// Service constants
	DefaultMaxLength      = 10
	DefaultTemperature    = 0.7
	DefaultMaxTokens       = 4096
	DefaultTopP            = 1.0
	DefaultN               = 1
	RequestTimeoutSeconds = 10
	TokenEstimateRatio    = 4
	SummarizeThreshold    = 300
)

// Error message constants
const (
	ErrorStreamUnsupported = "Streaming unsupported by client"
	ErrorNoContent         = "no content in answer"
	ErrorEndOfStream       = "End of stream reached"
	ErrorDoneBreak         = "DONE break"
)

// HTTP constants
const (
	ContentTypeJSON     = "application/json"
	AcceptEventStream   = "text/event-stream"
	CacheControlNoCache = "no-cache"
	ConnectionKeepAlive = "keep-alive"
)
