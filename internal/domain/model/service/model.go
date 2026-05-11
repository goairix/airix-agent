package service

import (
	"context"

	"github.com/google/uuid"

	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 模型管理领域服务接口
type Service interface {
	// Provider 管理
	CreateProvider(ctx context.Context, name string, protocol valueobject.Protocol, baseURL string, authType valueobject.AuthType, capabilities []valueobject.Capability) (*model.Provider, error)
	GetProviderByID(ctx context.Context, providerID uuid.UUID) (*model.Provider, error)
	ListProviders(ctx context.Context, pagination repository.Pagination) ([]model.Provider, int64, error)
	UpdateProvider(ctx context.Context, provider *model.Provider) error
	DeleteProvider(ctx context.Context, providerID uuid.UUID) error

	// Instance 管理
	CreateInstance(ctx context.Context, workspaceID, providerID uuid.UUID, modelName string, capability valueobject.Capability, apiKey string, parameters map[string]any, rateLimit model.RateLimit) (*model.Instance, error)
	GetInstanceByID(ctx context.Context, instanceID uuid.UUID) (*model.Instance, error)
	ListInstancesByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Instance, int64, error)
	ListInstancesByCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination repository.Pagination) ([]model.Instance, int64, error)
	UpdateInstance(ctx context.Context, instance *model.Instance) error
	DeleteInstance(ctx context.Context, instanceID uuid.UUID) error
	EnableInstance(ctx context.Context, instanceID uuid.UUID) error
	DisableInstance(ctx context.Context, instanceID uuid.UUID) error
}

type modelDomainService struct {
	baseTraceSpanName string
	providerRepo      repository.ProviderRepository
	instanceRepo      repository.InstanceRepository
}

func NewModelDomainService(providerRepo repository.ProviderRepository, instanceRepo repository.InstanceRepository) Service {
	return &modelDomainService{
		baseTraceSpanName: "domain.model.service.ModelDomainService",
		providerRepo:      providerRepo,
		instanceRepo:      instanceRepo,
	}
}

// --- Provider ---

func (svc *modelDomainService) CreateProvider(ctx context.Context, name string, protocol valueobject.Protocol, baseURL string, authType valueobject.AuthType, capabilities []valueobject.Capability) (*model.Provider, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateProvider")
	defer span.End()

	provider, err := model.NewProvider(name, protocol, baseURL, authType)
	if err != nil {
		return nil, err
	}
	for _, cap := range capabilities {
		provider.AddCapability(cap)
	}
	if err = svc.providerRepo.Save(spanCtx, provider); err != nil {
		return nil, modelErrors.ErrProviderSaveFailed.WrapNew(err)
	}
	return provider, nil
}

func (svc *modelDomainService) GetProviderByID(ctx context.Context, providerID uuid.UUID) (*model.Provider, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetProviderByID")
	defer span.End()

	provider, err := svc.providerRepo.FindByID(spanCtx, providerID)
	if err != nil {
		return nil, modelErrors.ErrProviderQueryFailed.WrapNew(err)
	}
	if provider == nil {
		return nil, modelErrors.ErrProviderNotFound
	}
	return provider, nil
}

func (svc *modelDomainService) ListProviders(ctx context.Context, pagination repository.Pagination) ([]model.Provider, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListProviders")
	defer span.End()

	providers, total, err := svc.providerRepo.FindAll(spanCtx, pagination)
	if err != nil {
		return nil, 0, modelErrors.ErrProviderQueryFailed.WrapNew(err)
	}
	return providers, total, nil
}

func (svc *modelDomainService) UpdateProvider(ctx context.Context, provider *model.Provider) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateProvider")
	defer span.End()

	if err := provider.Validate(); err != nil {
		return err
	}
	if err := svc.providerRepo.Save(spanCtx, provider); err != nil {
		return modelErrors.ErrProviderSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteProvider")
	defer span.End()

	if _, err := svc.GetProviderByID(spanCtx, providerID); err != nil {
		return err
	}
	exists, err := svc.instanceRepo.ExistsByProviderID(spanCtx, providerID)
	if err != nil {
		return modelErrors.ErrProviderDeleteFailed.WrapNew(err)
	}
	if exists {
		return modelErrors.ErrProviderHasInstances
	}
	if err = svc.providerRepo.Delete(spanCtx, providerID); err != nil {
		return modelErrors.ErrProviderDeleteFailed.WrapNew(err)
	}
	return nil
}

// --- Instance ---

func (svc *modelDomainService) CreateInstance(ctx context.Context, workspaceID, providerID uuid.UUID, modelName string, capability valueobject.Capability, apiKey string, parameters map[string]any, rateLimit model.RateLimit) (*model.Instance, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateInstance")
	defer span.End()

	if _, err := svc.GetProviderByID(spanCtx, providerID); err != nil {
		return nil, err
	}

	instance, err := model.NewInstance(workspaceID, providerID, modelName, capability)
	if err != nil {
		return nil, err
	}
	instance.SetAPIKey(apiKey)
	instance.Parameters = parameters
	instance.RateLimit = rateLimit

	if err = svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return nil, modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return instance, nil
}

func (svc *modelDomainService) GetInstanceByID(ctx context.Context, instanceID uuid.UUID) (*model.Instance, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetInstanceByID")
	defer span.End()

	instance, err := svc.instanceRepo.FindByID(spanCtx, instanceID)
	if err != nil {
		return nil, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	if instance == nil {
		return nil, modelErrors.ErrInstanceNotFound
	}
	return instance, nil
}

func (svc *modelDomainService) ListInstancesByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListInstancesByWorkspace")
	defer span.End()

	instances, total, err := svc.instanceRepo.FindByWorkspace(spanCtx, workspaceID, pagination)
	if err != nil {
		return nil, 0, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	return instances, total, nil
}

func (svc *modelDomainService) ListInstancesByCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination repository.Pagination) ([]model.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListInstancesByCapability")
	defer span.End()

	instances, total, err := svc.instanceRepo.FindByWorkspaceAndCapability(spanCtx, workspaceID, capability, pagination)
	if err != nil {
		return nil, 0, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	return instances, total, nil
}

func (svc *modelDomainService) UpdateInstance(ctx context.Context, instance *model.Instance) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateInstance")
	defer span.End()

	if err := instance.Validate(); err != nil {
		return err
	}
	if err := svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteInstance")
	defer span.End()

	if _, err := svc.GetInstanceByID(spanCtx, instanceID); err != nil {
		return err
	}
	if err := svc.instanceRepo.Delete(spanCtx, instanceID); err != nil {
		return modelErrors.ErrInstanceDeleteFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) EnableInstance(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".EnableInstance")
	defer span.End()

	instance, err := svc.GetInstanceByID(spanCtx, instanceID)
	if err != nil {
		return err
	}
	instance.Enable()
	if err = svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) DisableInstance(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DisableInstance")
	defer span.End()

	instance, err := svc.GetInstanceByID(spanCtx, instanceID)
	if err != nil {
		return err
	}
	instance.Disable()
	if err = svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return nil
}
