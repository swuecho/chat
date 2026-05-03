package dto

import "time"

// API and request constants
const (
	DefaultRequestTimeout      = 5 * time.Minute
	MaxStreamingLoopIterations = 10000
	SmallAnswerThreshold       = 200
	FlushCharacterThreshold    = 500
	TestPrefixLength           = 16
	DefaultPageSize            = 200
	MaxHistoryItems            = 10000
	DefaultPageLimit           = 30
	TestDemoPrefix             = "test_demo_bestqa"
	DefaultMaxLength           = 10
	DefaultTemperature         = 0.7
	DefaultMaxTokens           = 4096
	DefaultTopP                = 1.0
	DefaultN                   = 1
	RequestTimeoutSeconds      = 10
	TokenEstimateRatio         = 4
	SummarizeThreshold         = 300
	DefaultSystemPromptText    = "You are a helpful, concise assistant. Ask clarifying questions when needed. Provide accurate answers with short reasoning and actionable steps. If unsure, say so and suggest how to verify."
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
