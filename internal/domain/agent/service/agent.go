package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	agentErrors "github.com/dysodeng/app/internal/domain/agent/errors"
	"github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Agent 领域服务接口
type Service interface {
	Create(ctx context.Context, workspaceID uuid.UUID, name, description string, agentType valueobject.AgentType, createdBy uuid.UUID) (*model.Agent, error)
	GetByID(ctx context.Context, agentID uuid.UUID) (*model.Agent, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Agent, int64, error)
	Update(ctx context.Context, agent *model.Agent) error
	Delete(ctx context.Context, agentID uuid.UUID) error
	Publish(ctx context.Context, agentID, operatorID uuid.UUID, changeLog string, snapshot model.AgentSnapshot) (*model.AgentRelease, error)
	Rollback(ctx context.Context, agentID uuid.UUID, releaseID string) (*model.Agent, error)
	GetRelease(ctx context.Context, releaseID string) (*model.AgentRelease, error)
	GetActiveRelease(ctx context.Context, agentID uuid.UUID) (*model.AgentRelease, error)
	ListReleases(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]model.AgentRelease, int64, error)
}

type agentDomainService struct {
	baseTraceSpanName string
	repo              repository.Repository
	releaseRepo       repository.ReleaseRepository
}

func NewAgentDomainService(repo repository.Repository, releaseRepo repository.ReleaseRepository) Service {
	return &agentDomainService{
		baseTraceSpanName: "domain.agent.service.AgentDomainService",
		repo:              repo,
		releaseRepo:       releaseRepo,
	}
}

func (svc *agentDomainService) Create(ctx context.Context, workspaceID uuid.UUID, name, description string, agentType valueobject.AgentType, createdBy uuid.UUID) (*model.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Create")
	defer span.End()

	agent, err := model.NewAgent(workspaceID, name, description, agentType, createdBy)
	if err != nil {
		return nil, err
	}
	if err = svc.repo.Save(spanCtx, agent); err != nil {
		return nil, agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}
	return agent, nil
}

func (svc *agentDomainService) GetByID(ctx context.Context, agentID uuid.UUID) (*model.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetByID")
	defer span.End()

	agent, err := svc.repo.FindByID(spanCtx, agentID)
	if err != nil {
		return nil, agentErrors.ErrAgentQueryFailed.WrapNew(err)
	}
	if agent == nil {
		return nil, agentErrors.ErrAgentNotFound
	}
	return agent, nil
}

func (svc *agentDomainService) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Agent, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListByWorkspace")
	defer span.End()

	agents, total, err := svc.repo.FindByWorkspace(spanCtx, workspaceID, pagination)
	if err != nil {
		return nil, 0, agentErrors.ErrAgentQueryFailed.WrapNew(err)
	}
	return agents, total, nil
}

func (svc *agentDomainService) Update(ctx context.Context, agent *model.Agent) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Update")
	defer span.End()

	if err := agent.Validate(); err != nil {
		return err
	}
	if err := svc.repo.Save(spanCtx, agent); err != nil {
		return agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *agentDomainService) Delete(ctx context.Context, agentID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Delete")
	defer span.End()

	if _, err := svc.GetByID(spanCtx, agentID); err != nil {
		return err
	}
	if err := svc.repo.Delete(spanCtx, agentID); err != nil {
		return agentErrors.ErrAgentDeleteFailed.WrapNew(err)
	}
	return nil
}

func (svc *agentDomainService) Publish(ctx context.Context, agentID, operatorID uuid.UUID, changeLog string, snapshot model.AgentSnapshot) (*model.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Publish")
	defer span.End()

	agent, err := svc.GetByID(spanCtx, agentID)
	if err != nil {
		return nil, err
	}

	// 将旧 active Release 置为 inactive
	if err = svc.releaseRepo.DeactivateAll(spanCtx, agentID); err != nil {
		return nil, agentErrors.ErrReleaseSaveFailed.WrapNew(err)
	}

	releaseID := time.Now().Format("20060102-150405")
	release := model.NewAgentRelease(agentID, agent.WorkspaceID, operatorID, changeLog, snapshot)
	release.ReleaseID = releaseID
	release.ReleasedAt = time.Now()

	if err = svc.releaseRepo.Save(spanCtx, release); err != nil {
		return nil, agentErrors.ErrReleaseSaveFailed.WrapNew(err)
	}

	agent.ActiveReleaseID = releaseID
	agent.Status = valueobject.AgentStatusActive
	if err = svc.repo.Save(spanCtx, agent); err != nil {
		return nil, agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}

	return release, nil
}

func (svc *agentDomainService) Rollback(ctx context.Context, agentID uuid.UUID, releaseID string) (*model.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Rollback")
	defer span.End()

	agent, err := svc.GetByID(spanCtx, agentID)
	if err != nil {
		return nil, err
	}

	release, err := svc.releaseRepo.FindByID(spanCtx, releaseID)
	if err != nil {
		return nil, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	if release == nil {
		return nil, agentErrors.ErrReleaseNotFound
	}

	// 基于 Snapshot 覆写 Agent 草稿配置
	agent.Name = release.Snapshot.Name
	agent.Description = release.Snapshot.Description
	agent.AgentType = release.Snapshot.AgentType
	agent.SystemPrompt = release.Snapshot.SystemPrompt
	agent.ModelConfig = release.Snapshot.ModelConfig
	agent.MemoryConfig = release.Snapshot.MemoryConfig
	agent.CollaborationConfig = release.Snapshot.CollaborationConfig
	agent.SandboxConfig = release.Snapshot.SandboxConfig
	// 回滚不更改 ActiveReleaseID，也不自动发布
	agent.Status = valueobject.AgentStatusDraft

	if err = svc.repo.Save(spanCtx, agent); err != nil {
		return nil, agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}

	return agent, nil
}

func (svc *agentDomainService) GetRelease(ctx context.Context, releaseID string) (*model.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetRelease")
	defer span.End()

	release, err := svc.releaseRepo.FindByID(spanCtx, releaseID)
	if err != nil {
		return nil, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	if release == nil {
		return nil, agentErrors.ErrReleaseNotFound
	}
	return release, nil
}

func (svc *agentDomainService) GetActiveRelease(ctx context.Context, agentID uuid.UUID) (*model.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetActiveRelease")
	defer span.End()

	release, err := svc.releaseRepo.FindActive(spanCtx, agentID)
	if err != nil {
		return nil, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	if release == nil {
		return nil, agentErrors.ErrReleaseNotFound
	}
	return release, nil
}

func (svc *agentDomainService) ListReleases(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]model.AgentRelease, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListReleases")
	defer span.End()

	releases, total, err := svc.releaseRepo.ListByAgent(spanCtx, agentID, pagination)
	if err != nil {
		return nil, 0, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	return releases, total, nil
}
