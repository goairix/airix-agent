package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/session/valueobject"
)

// Message 一轮对话（用户 query → 最终回复）
type Message struct {
	ID                  uuid.UUID
	SessionID           uuid.UUID
	WorkspaceID         uuid.UUID
	AgentID             uuid.UUID
	SortOrder           int64 // 填充时间戳，(session_id, sort_order) 联合索引
	Query               string
	AgentInput          map[string]any // Agent 注入的变量参数
	Status              valueobject.MessageStatus
	TotalTokens         int64
	InputTokens         int64
	OutputTokens        int64
	CachedTokens        int64
	ExecutionTimeMs     int64
	FirstTokenLatencyMs int64
	CreatedAt           time.Time
	CompletedAt         *time.Time
}

func NewMessage(sessionID, workspaceID, agentID uuid.UUID, query string, sortOrder int64) *Message {
	id, _ := uuid.NewV7()
	return &Message{
		ID:          id,
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		AgentID:     agentID,
		SortOrder:   sortOrder,
		Query:       query,
		Status:      valueobject.MessageStatusRunning,
		CreatedAt:   time.Now(),
	}
}
