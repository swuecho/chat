// Package main — Re-exports commonly used constants from the dto package.
package main

import "github.com/swuecho/chat_backend/dto"

const (
	DefaultRequestTimeout      = dto.DefaultRequestTimeout
	MaxStreamingLoopIterations = dto.MaxStreamingLoopIterations
	SmallAnswerThreshold       = dto.SmallAnswerThreshold
	FlushCharacterThreshold    = dto.FlushCharacterThreshold
	TestPrefixLength           = dto.TestPrefixLength
	TestDemoPrefix             = dto.TestDemoPrefix
	DefaultMaxLength           = dto.DefaultMaxLength
	DefaultTemperature         = dto.DefaultTemperature
	DefaultMaxTokens           = dto.DefaultMaxTokens
	DefaultTopP                = dto.DefaultTopP
	DefaultN                   = dto.DefaultN
	RequestTimeoutSeconds      = dto.RequestTimeoutSeconds
	TokenEstimateRatio         = dto.TokenEstimateRatio
	SummarizeThreshold         = dto.SummarizeThreshold
	DefaultSystemPromptText    = dto.DefaultSystemPromptText
	DefaultPageSize            = dto.DefaultPageSize
	MaxHistoryItems            = dto.MaxHistoryItems
	DefaultPageLimit           = dto.DefaultPageLimit
	ErrorStreamUnsupported     = dto.ErrorStreamUnsupported
	ErrorNoContent             = dto.ErrorNoContent
	ErrorEndOfStream           = dto.ErrorEndOfStream
	ErrorDoneBreak             = dto.ErrorDoneBreak
	ContentTypeJSON            = dto.ContentTypeJSON
	AcceptEventStream          = dto.AcceptEventStream
	CacheControlNoCache        = dto.CacheControlNoCache
	ConnectionKeepAlive        = dto.ConnectionKeepAlive
)
