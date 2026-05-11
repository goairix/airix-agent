package model

import (
	"context"
	"fmt"

	openaiEmbedding "github.com/cloudwego/eino-ext/components/embedding/openai"
	claudeModel "github.com/cloudwego/eino-ext/components/model/claude"
	geminiModel "github.com/cloudwego/eino-ext/components/model/gemini"
	openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	einoModel "github.com/cloudwego/eino/components/model"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// AdapterFactory 协议适配工厂
type AdapterFactory struct{}

func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{}
}

// CreateChatModel 根据 Provider 协议和 Instance 配置创建 Eino ToolCallingChatModel
func (f *AdapterFactory) CreateChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	switch provider.Protocol {
	case valueobject.ProtocolOpenAICompatible:
		return f.createOpenAIChatModel(ctx, provider, instance)
	case valueobject.ProtocolAnthropic:
		return f.createAnthropicChatModel(ctx, provider, instance)
	case valueobject.ProtocolGoogle:
		return f.createGoogleChatModel(ctx, provider, instance)
	case valueobject.ProtocolCustom:
		return f.createCustomChatModel(ctx, provider, instance)
	default:
		return nil, fmt.Errorf("不支持的协议类型: %s", provider.Protocol.String())
	}
}

// CreateEmbedder 根据 Provider 协议和 Instance 配置创建 Eino Embedder
func (f *AdapterFactory) CreateEmbedder(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (embedding.Embedder, error) {
	switch provider.Protocol {
	case valueobject.ProtocolOpenAICompatible:
		return f.createOpenAIEmbedder(ctx, provider, instance)
	default:
		return nil, fmt.Errorf("不支持的 Embedding 协议类型: %s", provider.Protocol.String())
	}
}

// CreateReranker 根据 Provider 协议和 Instance 配置创建 document.Transformer 接口的 Reranker
func (f *AdapterFactory) CreateReranker(_ context.Context, _ *domainModel.Provider, _ *domainModel.Instance) (document.Transformer, error) {
	return nil, fmt.Errorf("Reranker 适配器待实现")
}

func (f *AdapterFactory) createOpenAIChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	cfg := &openaiModel.ChatModelConfig{
		Model:  instance.ModelName,
		APIKey: instance.APIKey,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = provider.BaseURL
	}
	if temp, ok := instance.Parameters["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			t := float32(v)
			cfg.Temperature = &t
		}
	}
	if maxTokens, ok := instance.Parameters["max_tokens"]; ok {
		if v, ok := maxTokens.(float64); ok {
			mt := int(v)
			cfg.MaxCompletionTokens = &mt
		}
	}
	return openaiModel.NewChatModel(ctx, cfg)
}

func (f *AdapterFactory) createAnthropicChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	cfg := &claudeModel.Config{
		Model:     instance.ModelName,
		APIKey:    instance.APIKey,
		MaxTokens: 4096,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = &provider.BaseURL
	}
	if maxTokens, ok := instance.Parameters["max_tokens"]; ok {
		if v, ok := maxTokens.(float64); ok {
			cfg.MaxTokens = int(v)
		}
	}
	if temp, ok := instance.Parameters["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			t := float32(v)
			cfg.Temperature = &t
		}
	}
	return claudeModel.NewChatModel(ctx, cfg)
}

func (f *AdapterFactory) createGoogleChatModel(_ context.Context, _ *domainModel.Provider, _ *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	_ = &geminiModel.Config{}
	return nil, fmt.Errorf("Google Gemini ChatModel 适配器待完善（需创建 genai.Client）")
}

func (f *AdapterFactory) createCustomChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	cfg := &openaiModel.ChatModelConfig{
		Model:  instance.ModelName,
		APIKey: instance.APIKey,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = provider.BaseURL
	}
	if temp, ok := instance.Parameters["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			t := float32(v)
			cfg.Temperature = &t
		}
	}
	if maxTokens, ok := instance.Parameters["max_tokens"]; ok {
		if v, ok := maxTokens.(float64); ok {
			mt := int(v)
			cfg.MaxCompletionTokens = &mt
		}
	}
	return openaiModel.NewChatModel(ctx, cfg)
}

func (f *AdapterFactory) createOpenAIEmbedder(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (embedding.Embedder, error) {
	cfg := &openaiEmbedding.EmbeddingConfig{
		Model:  instance.ModelName,
		APIKey: instance.APIKey,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = provider.BaseURL
	}
	return openaiEmbedding.NewEmbedder(ctx, cfg)
}
