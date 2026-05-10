package service

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/session/model"
	"github.com/dysodeng/app/internal/domain/agent/session/repository"
	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
)

// ContextAssembler 上下文组装接口（应用层）
type ContextAssembler interface {
	// Assemble 组装指定 session 的 LLM 上下文消息列表
	Assemble(ctx context.Context, sessionID uuid.UUID) ([]*schema.Message, error)
}

// SlidingWindowAssembler 滑动窗口上下文组装器
// 注意：windowSize 来自 Agent 配置，不通过 Wire 注入；由应用服务在运行时按 Agent 配置动态构造。
type SlidingWindowAssembler struct {
	windowSize  int
	messageRepo repository.MessageRepository
	itemRepo    repository.MessageItemRepository
}

func NewSlidingWindowAssembler(
	windowSize int,
	messageRepo repository.MessageRepository,
	itemRepo repository.MessageItemRepository,
) *SlidingWindowAssembler {
	return &SlidingWindowAssembler{
		windowSize:  windowSize,
		messageRepo: messageRepo,
		itemRepo:    itemRepo,
	}
}

func (a *SlidingWindowAssembler) Assemble(ctx context.Context, sessionID uuid.UUID) ([]*schema.Message, error) {
	messages, err := a.messageRepo.GetLatestN(ctx, sessionID, a.windowSize)
	if err != nil {
		return nil, err
	}
	var result []*schema.Message
	for _, msg := range messages {
		result = append(result, schema.UserMessage(msg.Query))
		items, err := a.itemRepo.ListByMessage(ctx, msg.ID)
		if err != nil {
			return nil, err
		}
		sort.Sort(model.ByOrder(items))
		result = append(result, MapItemsToMessages(items)...)
	}
	return result, nil
}

// MapItemsToMessages 将 MessageItem 列表映射为 Eino schema.Message 列表。
// error 类型不注入 LLM 上下文，tool_call 拆为两条消息。
func MapItemsToMessages(items []*model.MessageItem) []*schema.Message {
	var msgs []*schema.Message
	for _, item := range items {
		switch item.ItemType {
		case valueobject.MessageItemTypeThinking:
			msgs = append(msgs, schema.AssistantMessage(item.Content.Text, nil))
		case valueobject.MessageItemTypeAssistant:
			msgs = append(msgs, schema.AssistantMessage(item.Content.Text, nil))
		case valueobject.MessageItemTypeToolCall:
			toolCallJSON, _ := json.Marshal(item.Content.Arguments)
			assistantMsg := &schema.Message{
				Role: schema.Assistant,
				ToolCalls: []schema.ToolCall{
					{
						ID:   item.Content.ToolCallID,
						Type: "function",
						Function: schema.FunctionCall{
							Name:      item.Content.ToolName,
							Arguments: string(toolCallJSON),
						},
					},
				},
			}
			msgs = append(msgs, assistantMsg)
			resultJSON, _ := json.Marshal(item.Content.Result)
			if item.Content.Error != "" {
				resultJSON, _ = json.Marshal(map[string]string{"error": item.Content.Error})
			}
			toolMsg := schema.ToolMessage(string(resultJSON), item.Content.ToolCallID)
			msgs = append(msgs, toolMsg)
		case valueobject.MessageItemTypeError:
			// error 不注入 LLM 上下文
		}
	}
	return msgs
}
