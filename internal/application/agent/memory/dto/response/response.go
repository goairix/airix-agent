package response

import "time"

// MemoryResponse 记忆响应
type MemoryResponse struct {
	MemoryID    string    `json:"memory_id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	AgentID     string    `json:"agent_id"`
	MemoryType  string    `json:"memory_type"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Importance  float64   `json:"importance"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}
