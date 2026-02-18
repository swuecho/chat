package dto

// UpdateChatSessionRequest represents a request to update a chat session
type UpdateChatSessionRequest struct {
	Uuid               string  `json:"uuid"`
	Topic              string  `json:"topic"`
	MaxLength          int32   `json:"maxLength"`
	Temperature        float64 `json:"temperature"`
	Model              string  `json:"model"`
	TopP               float64 `json:"topP"`
	N                  int32   `json:"n"`
	MaxTokens          int32   `json:"maxTokens"`
	Debug              bool    `json:"debug"`
	SummarizeMode      bool    `json:"summarizeMode"`
	CodeRunnerEnabled  bool    `json:"codeRunnerEnabled"`
	ArtifactEnabled   bool    `json:"artifactEnabled"`
	ExploreMode        bool    `json:"exploreMode"`
	WorkspaceUUID      string  `json:"workspaceUuid,omitempty"`
}
