package domain

// Artifact is structured content extracted from an assistant message. It is an
// application concept shared by persistence and transport adapters.
type Artifact struct {
	UUID     string `json:"uuid"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}
