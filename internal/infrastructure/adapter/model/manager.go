package model

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	einoModel "github.com/cloudwego/eino/components/model"

	appPort "github.com/dysodeng/app/internal/application/model/port"
	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Manager 模型管理器
type Manager struct {
	baseTraceSpanName string
	providerRepo      repository.ProviderRepository
	instanceRepo      repository.InstanceRepository
	factory           *AdapterFactory
}

func NewManager(providerRepo repository.ProviderRepository, instanceRepo repository.InstanceRepository, factory *AdapterFactory) appPort.Manager {
	return &Manager{
		baseTraceSpanName: "infrastructure.adapter.model.Manager",
		providerRepo:      providerRepo,
		instanceRepo:      instanceRepo,
		factory:           factory,
	}
}

func (m *Manager) GetChatModel(ctx context.Context, instanceID string) (einoModel.ToolCallingChatModel, error) {
	spanCtx, span := trace.Tracer().Start(ctx, m.baseTraceSpanName+".GetChatModel")
	defer span.End()

	instance, provider, err := m.loadInstanceAndProvider(spanCtx, instanceID)
	if err != nil {
		return nil, err
	}

	if instance.Capability != valueobject.CapabilityChat {
		return nil, errors.New("该模型实例不支持 Chat 能力")
	}

	return m.factory.CreateChatModel(spanCtx, provider, instance)
}

func (m *Manager) GetEmbedder(ctx context.Context, instanceID string) (embedding.Embedder, error) {
	spanCtx, span := trace.Tracer().Start(ctx, m.baseTraceSpanName+".GetEmbedder")
	defer span.End()

	instance, provider, err := m.loadInstanceAndProvider(spanCtx, instanceID)
	if err != nil {
		return nil, err
	}

	if instance.Capability != valueobject.CapabilityEmbedding {
		return nil, errors.New("该模型实例不支持 Embedding 能力")
	}

	return m.factory.CreateEmbedder(spanCtx, provider, instance)
}

func (m *Manager) GetReranker(ctx context.Context, instanceID string) (document.Transformer, error) {
	spanCtx, span := trace.Tracer().Start(ctx, m.baseTraceSpanName+".GetReranker")
	defer span.End()

	instance, provider, err := m.loadInstanceAndProvider(spanCtx, instanceID)
	if err != nil {
		return nil, err
	}

	if instance.Capability != valueobject.CapabilityRerank {
		return nil, errors.New("该模型实例不支持 Rerank 能力")
	}

	return m.factory.CreateReranker(spanCtx, provider, instance)
}

func (m *Manager) loadInstanceAndProvider(ctx context.Context, instanceID string) (*domainModel.Instance, *domainModel.Provider, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, nil, errors.New("模型实例 ID 格式错误")
	}

	instance, err := m.instanceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	if instance == nil {
		return nil, nil, modelErrors.ErrInstanceNotFound
	}
	if !instance.IsActive() {
		return nil, nil, modelErrors.ErrInstanceDisabled
	}

	provider, err := m.providerRepo.FindByID(ctx, instance.ProviderID)
	if err != nil {
		return nil, nil, modelErrors.ErrProviderQueryFailed.WrapNew(err)
	}
	if provider == nil {
		return nil, nil, modelErrors.ErrProviderNotFound
	}

	return instance, provider, nil
}
