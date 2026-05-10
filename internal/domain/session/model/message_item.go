package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/session/valueobject"
)

// MessageItemContent 步骤内容（统一 JSON 结构）
type MessageItemContent struct {
	Text       string         `json:"text,omitempty"`         // thinking / assistant
	ToolName   string         `json:"tool_name,omitempty"`    // tool_call
	ToolCallID string         `json:"tool_call_id,omitempty"` // tool_call
	Arguments  map[string]any `json:"arguments,omitempty"`    // tool_call 入参
	Result     map[string]any `json:"result,omitempty"`       // tool_call 结果
	Error      string         `json:"error,omitempty"`        // tool_call 错误 / error 类型消息
	Code       string         `json:"code,omitempty"`         // error 类型错误码
}

// MessageItem 轮次内单个步骤
type MessageItem struct {
	ID           uuid.UUID
	MessageID    uuid.UUID
	SessionID    uuid.UUID
	SortOrder    int // 应用层自增（从 0 开始），无索引
	ItemType     valueobject.MessageItemType
	IsFinal      bool // 是否为最终回复（ItemType=assistant 时有意义）
	Content      MessageItemContent
	InputTokens  int64
	OutputTokens int64
	LatencyMs    int64
	CreatedAt    time.Time
}

// ByOrder 实现 sort.Interface，按 SortOrder 排序
type ByOrder []*MessageItem

func (a ByOrder) Len() int           { return len(a) }
func (a ByOrder) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOrder) Less(i, j int) bool { return a[i].SortOrder < a[j].SortOrder }
