package response

import "time"

// SessionResponse 会话响应
type SessionResponse struct {
	SessionID   string    `json:"session_id"`
	WorkspaceID string    `json:"workspace_id"`
	AgentID     string    `json:"agent_id"`
	ReleaseID   string    `json:"release_id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// MessageResponse 消息（轮次）响应
type MessageResponse struct {
	MessageID   string    `json:"message_id"`
	SessionID   string    `json:"session_id"`
	Query       string    `json:"query"`
	Status      string    `json:"status"`
	TotalTokens int64     `json:"total_tokens"`
	CreatedAt   time.Time `json:"created_at"`
}
