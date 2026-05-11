package port

import (
	"context"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	einoModel "github.com/cloudwego/eino/components/model"
)

// Manager 模型管理器接口
// Agent 运行时通过此接口获取 Eino ChatModel / Embedder / Reranker 实例
type Manager interface {
	// GetChatModel 根据模型实例 ID 获取 Eino ToolCallingChatModel
	GetChatModel(ctx context.Context, instanceID string) (einoModel.ToolCallingChatModel, error)
	// GetEmbedder 根据模型实例 ID 获取 Eino Embedder
	GetEmbedder(ctx context.Context, instanceID string) (embedding.Embedder, error)
	// GetReranker 根据模型实例 ID 获取基于 document.Transformer 接口的 Reranker
	GetReranker(ctx context.Context, instanceID string) (document.Transformer, error)
}
