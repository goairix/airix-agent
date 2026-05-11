package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/dysodeng/app/internal/application/model/dto/command"
	"github.com/dysodeng/app/internal/application/model/dto/response"
	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	domainService "github.com/dysodeng/app/internal/domain/model/service"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 模型管理应用服务接口
type Service interface {
	// Provider 管理
	CreateProvider(ctx context.Context, cmd *command.CreateProviderCommand) (*response.ProviderResponse, error)
	GetProvider(ctx context.Context, providerID string) (*response.ProviderResponse, error)
	ListProviders(ctx context.Context, page, pageSize int) (*response.ProviderListResponse, error)
	UpdateProvider(ctx context.Context, cmd *command.UpdateProviderCommand) (*response.ProviderResponse, error)
	DeleteProvider(ctx context.Context, providerID string) error

	// Instance 管理
	CreateInstance(ctx context.Context, cmd *command.CreateInstanceCommand) (*response.InstanceResponse, error)
	GetInstance(ctx context.Context, instanceID string) (*response.InstanceResponse, error)
	ListInstances(ctx context.Context, workspaceID string, page, pageSize int) (*response.InstanceListResponse, error)
	UpdateInstance(ctx context.Context, cmd *command.UpdateInstanceCommand) (*response.InstanceResponse, error)
	DeleteInstance(ctx context.Context, instanceID string) error
	EnableInstance(ctx context.Context, instanceID string) error
	DisableInstance(ctx context.Context, instanceID string) error
}

type modelApplicationService struct {
	baseTraceSpanName string
	domainService     domainService.Service
}

func NewModelApplicationService(domainSvc domainService.Service) Service {
	return &modelApplicationService{
		baseTraceSpanName: "application.model.service.ModelApplicationService",
		domainService:     domainSvc,
	}
}

func (svc *modelApplicationService) CreateProvider(ctx context.Context, cmd *command.CreateProviderCommand) (*response.ProviderResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateProvider")
	defer span.End()

	caps := lo.Map(cmd.Capabilities, func(c uint8, _ int) valueobject.Capability {
		return valueobject.Capability(c)
	})

	provider, err := svc.domainService.CreateProvider(spanCtx, cmd.Name, valueobject.Protocol(cmd.Protocol), cmd.BaseURL, valueobject.AuthType(cmd.AuthType), caps)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toProviderResponse(provider), nil
}

func (svc *modelApplicationService) GetProvider(ctx context.Context, providerID string) (*response.ProviderResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetProvider")
	defer span.End()

	id, err := uuid.Parse(providerID)
	if err != nil {
		return nil, errors.New("模型厂商 ID 格式错误")
	}
	provider, err := svc.domainService.GetProviderByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toProviderResponse(provider), nil
}

func (svc *modelApplicationService) ListProviders(ctx context.Context, page, pageSize int) (*response.ProviderListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListProviders")
	defer span.End()

	providers, total, err := svc.domainService.ListProviders(spanCtx, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := lo.Map(providers, func(p domainModel.Provider, _ int) response.ProviderResponse {
		return *toProviderResponse(&p)
	})
	return &response.ProviderListResponse{Record: records, Total: total}, nil
}

func (svc *modelApplicationService) UpdateProvider(ctx context.Context, cmd *command.UpdateProviderCommand) (*response.ProviderResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateProvider")
	defer span.End()

	id, err := uuid.Parse(cmd.ProviderID)
	if err != nil {
		return nil, errors.New("模型厂商 ID 格式错误")
	}
	provider, err := svc.domainService.GetProviderByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	provider.Name = cmd.Name
	provider.Protocol = valueobject.Protocol(cmd.Protocol)
	provider.BaseURL = cmd.BaseURL
	provider.AuthType = valueobject.AuthType(cmd.AuthType)
	provider.SupportedCapabilities = lo.Map(cmd.Capabilities, func(c uint8, _ int) valueobject.Capability {
		return valueobject.Capability(c)
	})

	if err = svc.domainService.UpdateProvider(spanCtx, provider); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toProviderResponse(provider), nil
}

func (svc *modelApplicationService) DeleteProvider(ctx context.Context, providerID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteProvider")
	defer span.End()

	id, err := uuid.Parse(providerID)
	if err != nil {
		return errors.New("模型厂商 ID 格式错误")
	}
	if err = svc.domainService.DeleteProvider(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *modelApplicationService) CreateInstance(ctx context.Context, cmd *command.CreateInstanceCommand) (*response.InstanceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateInstance")
	defer span.End()

	wsID, err := uuid.Parse(cmd.WorkspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	providerID, err := uuid.Parse(cmd.ProviderID)
	if err != nil {
		return nil, errors.New("模型厂商 ID 格式错误")
	}

	instance, err := svc.domainService.CreateInstance(spanCtx, wsID, providerID, cmd.ModelName, valueobject.Capability(cmd.Capability), cmd.APIKey, cmd.Parameters, domainModel.RateLimit{RPM: cmd.RateLimitRPM, TPM: cmd.RateLimitTPM})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toInstanceResponse(instance), nil
}

func (svc *modelApplicationService) GetInstance(ctx context.Context, instanceID string) (*response.InstanceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, errors.New("模型实例 ID 格式错误")
	}
	instance, err := svc.domainService.GetInstanceByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toInstanceResponse(instance), nil
}

func (svc *modelApplicationService) ListInstances(ctx context.Context, workspaceID string, page, pageSize int) (*response.InstanceListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListInstances")
	defer span.End()

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	instances, total, err := svc.domainService.ListInstancesByWorkspace(spanCtx, wsID, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := lo.Map(instances, func(inst domainModel.Instance, _ int) response.InstanceResponse {
		return *toInstanceResponse(&inst)
	})
	return &response.InstanceListResponse{Record: records, Total: total}, nil
}

func (svc *modelApplicationService) UpdateInstance(ctx context.Context, cmd *command.UpdateInstanceCommand) (*response.InstanceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateInstance")
	defer span.End()

	id, err := uuid.Parse(cmd.InstanceID)
	if err != nil {
		return nil, errors.New("模型实例 ID 格式错误")
	}
	instance, err := svc.domainService.GetInstanceByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	instance.ModelName = cmd.ModelName
	instance.Capability = valueobject.Capability(cmd.Capability)
	if cmd.APIKey != "" {
		instance.SetAPIKey(cmd.APIKey)
	}
	instance.Parameters = cmd.Parameters
	instance.RateLimit = domainModel.RateLimit{RPM: cmd.RateLimitRPM, TPM: cmd.RateLimitTPM}

	if err = svc.domainService.UpdateInstance(spanCtx, instance); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toInstanceResponse(instance), nil
}

func (svc *modelApplicationService) DeleteInstance(ctx context.Context, instanceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return errors.New("模型实例 ID 格式错误")
	}
	if err = svc.domainService.DeleteInstance(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *modelApplicationService) EnableInstance(ctx context.Context, instanceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".EnableInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return errors.New("模型实例 ID 格式错误")
	}
	if err = svc.domainService.EnableInstance(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *modelApplicationService) DisableInstance(ctx context.Context, instanceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DisableInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return errors.New("模型实例 ID 格式错误")
	}
	if err = svc.domainService.DisableInstance(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func toProviderResponse(p *domainModel.Provider) *response.ProviderResponse {
	caps := lo.Map(p.SupportedCapabilities, func(c valueobject.Capability, _ int) string {
		return c.String()
	})
	return &response.ProviderResponse{
		ProviderID:   p.ID.String(),
		Name:         p.Name,
		Protocol:     p.Protocol.String(),
		BaseURL:      p.BaseURL,
		AuthType:     p.AuthType.String(),
		Capabilities: caps,
		CreatedAt:    p.CreatedAt,
	}
}

func toInstanceResponse(inst *domainModel.Instance) *response.InstanceResponse {
	return &response.InstanceResponse{
		InstanceID:   inst.ID.String(),
		WorkspaceID:  inst.WorkspaceID.String(),
		ProviderID:   inst.ProviderID.String(),
		ModelName:    inst.ModelName,
		Capability:   inst.Capability.String(),
		Parameters:   inst.Parameters,
		RateLimitRPM: inst.RateLimit.RPM,
		RateLimitTPM: inst.RateLimit.TPM,
		Status:       inst.Status.String(),
		CreatedAt:    inst.CreatedAt,
	}
}
