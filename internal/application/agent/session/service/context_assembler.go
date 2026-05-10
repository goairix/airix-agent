package service

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/session/model"
	"github.com/dysodeng/app/internal/domain/agent/session/repository"
	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
)

// ContextAssembler 上下文组装接口（应用层）
type ContextAssembler interface {
	// Assemble 组装指定 session 的 LLM 上下文消息列表
	Assemble(ctx context.Context, sessionID uuid.UUID) ([]*schema.Message, error)
}

// ContextAssemblerFactory 上下文组装器工厂
// 根据 Agent 的 MemoryConfig 选择组装策略。共享依赖通过 DI 注入，
// 具体组装器的策略参数由 Agent 配置驱动。
type ContextAssemblerFactory struct {
	messageRepo repository.MessageRepository
	itemRepo    repository.MessageItemRepository
}

func NewContextAssemblerFactory(
	messageRepo repository.MessageRepository,
	itemRepo repository.MessageItemRepository,
) *ContextAssemblerFactory {
	return &ContextAssemblerFactory{
		messageRepo: messageRepo,
		itemRepo:    itemRepo,
	}
}

// Create 根据 Agent 的 MemoryConfig 返回对应的上下文组装器
func (f *ContextAssemblerFactory) Create(cfg agentModel.MemoryConfig) ContextAssembler {
	switch cfg.ContextMode {
	case "sliding_window":
		return NewSlidingWindowAssembler(cfg.ContextWindowSize, f.messageRepo, f.itemRepo)
	default:
		// 默认兜底：滑动窗口
		return NewSlidingWindowAssembler(cfg.ContextWindowSize, f.messageRepo, f.itemRepo)
	}
}

// SlidingWindowAssembler 滑动窗口上下文组装器
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
