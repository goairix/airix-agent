package command

// CreateSessionCommand 创建会话命令
type CreateSessionCommand struct {
	WorkspaceID string
	AgentID     string
	ReleaseID   string // 可选，不填则使用 active 版本
	UserID      string
	Title       string
}

// CompleteMessageCommand 完成一轮消息命令
type CompleteMessageCommand struct {
	MessageID           string
	TotalTokens         int64
	InputTokens         int64
	OutputTokens        int64
	CachedTokens        int64
	ExecutionTimeMs     int64
	FirstTokenLatencyMs int64
}
