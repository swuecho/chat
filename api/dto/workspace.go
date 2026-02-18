package dto

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	IsDefault   bool   `json:"isDefault"`
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

// UpdateWorkspaceOrderRequest represents a request to update workspace order
type UpdateWorkspaceOrderRequest struct {
	OrderPosition int32 `json:"orderPosition"`
}

// WorkspaceResponse represents a workspace in API responses
type WorkspaceResponse struct {
	Uuid          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	Icon          string `json:"icon"`
	IsDefault     bool   `json:"isDefault"`
	OrderPosition int32  `json:"orderPosition"`
	SessionCount  int64  `json:"sessionCount,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}
