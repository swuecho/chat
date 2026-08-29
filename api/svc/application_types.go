package svc

const (
	tokenEstimateRatio      = 4
	summarizeThreshold      = 300
	requestTimeoutSeconds   = 10
	defaultMaxLength        = 10
	defaultSystemPromptText = "You are a helpful, concise assistant. Ask clarifying questions when needed. Provide accurate answers with short reasoning and actionable steps. If unsure, say so and suggest how to verify."
)

// PageRequest describes application pagination independently of an HTTP
// request or response shape.
type PageRequest struct {
	Page int32
	Size int32
}

func (p PageRequest) Offset() int32 { return (p.Page - 1) * p.Size }

// PageWindow represents an already-resolved database paging window.
type PageWindow struct {
	Limit  int32
	Offset int32
}

// SimpleChatSession is the session-list read model returned by the application.
type SimpleChatSession struct {
	UUID            string
	Title           string
	MaxLength       int
	Temperature     float64
	TopP            float64
	N               int32
	MaxTokens       int32
	Debug           bool
	Model           string
	SummarizeMode   bool
	ArtifactEnabled bool
	WorkspaceUUID   string
}
