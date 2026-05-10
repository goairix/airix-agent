package strategy_test

import (
	"sort"
	"testing"

	"github.com/dysodeng/app/internal/application/agent/session/strategy"
	"github.com/dysodeng/app/internal/domain/agent/session/model"
	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
)

func TestMapItemsToMessages_ThinkingAndAssistant(t *testing.T) {
	items := []*model.MessageItem{
		{SortOrder: 1, ItemType: valueobject.MessageItemTypeAssistant, Content: model.MessageItemContent{Text: "hello"}, IsFinal: true},
		{SortOrder: 0, ItemType: valueobject.MessageItemTypeThinking, Content: model.MessageItemContent{Text: "thinking..."}},
	}
	sort.Sort(model.ByOrder(items))
	msgs := strategy.MapItemsToMessages(items)
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestMapItemsToMessages_ToolCall(t *testing.T) {
	items := []*model.MessageItem{
		{SortOrder: 0, ItemType: valueobject.MessageItemTypeToolCall, Content: model.MessageItemContent{
			ToolName:   "search",
			ToolCallID: "call_001",
			Arguments:  map[string]any{"q": "test"},
			Result:     map[string]any{"answer": "42"},
		}},
	}
	msgs := strategy.MapItemsToMessages(items)
	// tool_call 拆成两条: assistant(tool_calls) + tool(result)
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages for tool_call, got %d", len(msgs))
	}
}