package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/application/agent/dto/command"
	"github.com/dysodeng/app/internal/application/agent/dto/response"
	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	domainService "github.com/dysodeng/app/internal/domain/agent/service"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Agent 应用服务接口
type Service interface {
	CreateAgent(ctx context.Context, cmd *command.CreateAgentCommand) (*response.AgentResponse, error)
	GetAgent(ctx context.Context, agentID string) (*response.AgentResponse, error)
	ListAgents(ctx context.Context, workspaceID string, page, pageSize int) (*response.AgentListResponse, error)
	UpdateAgent(ctx context.Context, cmd *command.UpdateAgentCommand) (*response.AgentResponse, error)
	DeleteAgent(ctx context.Context, agentID string) error
	PublishAgent(ctx context.Context, cmd *command.PublishAgentCommand) (*response.AgentReleaseResponse, error)
	RollbackAgent(ctx context.Context, cmd *command.RollbackAgentCommand) (*response.AgentResponse, error)
	GetRelease(ctx context.Context, releaseID string) (*response.AgentReleaseResponse, error)
	ListReleases(ctx context.Context, agentID string, page, pageSize int) (*response.AgentReleaseListResponse, error)
}

type agentApplicationService struct {
	baseTraceSpanName string
	domainService     domainService.Service
}

func NewAgentApplicationService(domainSvc domainService.Service) Service {
	return &agentApplicationService{
		baseTraceSpanName: "application.agent.service.AgentApplicationService",
		domainService:     domainSvc,
	}
}

func (svc *agentApplicationService) CreateAgent(ctx context.Context, cmd *command.CreateAgentCommand) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateAgent")
	defer span.End()

	workspaceID, err := uuid.Parse(cmd.WorkspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	createdBy, err := uuid.Parse(cmd.CreatedBy)
	if err != nil {
		return nil, errors.New("创建人 ID 格式错误")
	}

	agent, err := svc.domainService.Create(
		spanCtx,
		workspaceID,
		cmd.Name,
		cmd.Description,
		valueobject.AgentType(cmd.AgentType),
		createdBy,
	)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) GetAgent(ctx context.Context, agentID string) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetAgent")
	defer span.End()

	id, err := uuid.Parse(agentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	agent, err := svc.domainService.GetByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) ListAgents(ctx context.Context, workspaceID string, page, pageSize int) (*response.AgentListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListAgents")
	defer span.End()

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}

	agents, total, err := svc.domainService.ListByWorkspace(spanCtx, wsID, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := make([]response.AgentResponse, 0, len(agents))
	for _, a := range agents {
		records = append(records, *toAgentResponse(&a))
	}
	return &response.AgentListResponse{Record: records, Total: total}, nil
}

func (svc *agentApplicationService) UpdateAgent(ctx context.Context, cmd *command.UpdateAgentCommand) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateAgent")
	defer span.End()

	id, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	agent, err := svc.domainService.GetByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	agent.Name = cmd.Name
	agent.Description = cmd.Description
	agent.SystemPrompt = cmd.SystemPrompt
	agent.ToolBindings = cmd.ToolBindings
	agent.KnowledgeBindings = cmd.KnowledgeBindings
	agent.SkillBindings = cmd.SkillBindings
	agent.MCPBindings = cmd.MCPBindings
	agent.ModelConfig = agentModel.ModelConfig{
		ModelInstanceID: cmd.ModelConfig.ModelInstanceID,
		Parameters:      cmd.ModelConfig.Parameters,
	}
	agent.MemoryConfig = agentModel.MemoryConfig{
		MemoryDriverType:  cmd.MemoryConfig.MemoryDriverType,
		ContextMode:       cmd.MemoryConfig.ContextMode,
		ContextWindowSize: cmd.MemoryConfig.ContextWindowSize,
		SummarizationConfig: agentModel.SummarizationConfig{
			SummaryModelInstanceID: cmd.MemoryConfig.SummarizationConfig.SummaryModelInstanceID,
			TriggerTokenThreshold:  cmd.MemoryConfig.SummarizationConfig.TriggerTokenThreshold,
		},
		GlobalMemoryMode: cmd.MemoryConfig.GlobalMemoryMode,
	}
	agent.CollaborationConfig = agentModel.CollaborationConfig{
		SubAgentIDs:        cmd.CollaborationConfig.SubAgentIDs,
		TransferPolicy:     cmd.CollaborationConfig.TransferPolicy,
		MaxDelegationDepth: cmd.CollaborationConfig.MaxDelegationDepth,
	}
	agent.SandboxConfig = agentModel.SandboxConfig{
		Enabled:     cmd.SandboxConfig.Enabled,
		SandboxType: cmd.SandboxConfig.SandboxType,
	}

	if err = svc.domainService.Update(spanCtx, agent); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) DeleteAgent(ctx context.Context, agentID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteAgent")
	defer span.End()

	id, err := uuid.Parse(agentID)
	if err != nil {
		return errors.New("Agent ID 格式错误")
	}
	if err = svc.domainService.Delete(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *agentApplicationService) PublishAgent(ctx context.Context, cmd *command.PublishAgentCommand) (*response.AgentReleaseResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".PublishAgent")
	defer span.End()

	agentID, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	operatorID, err := uuid.Parse(cmd.OperatorID)
	if err != nil {
		return nil, errors.New("操作人 ID 格式错误")
	}

	// 读取当前 Agent 的完整配置，构建快照
	agent, err := svc.domainService.GetByID(spanCtx, agentID)
	if err != nil {
		return nil, err
	}
	snapshot := agentModel.AgentSnapshot{
		Name:                agent.Name,
		Description:         agent.Description,
		AgentType:           agent.AgentType,
		SystemPrompt:        agent.SystemPrompt,
		ModelConfig:         agent.ModelConfig,
		MemoryConfig:        agent.MemoryConfig,
		CollaborationConfig: agent.CollaborationConfig,
		SandboxConfig:       agent.SandboxConfig,
	}

	release, err := svc.domainService.Publish(spanCtx, agentID, operatorID, cmd.ChangeLog, snapshot)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toReleaseResponse(release), nil
}

func (svc *agentApplicationService) RollbackAgent(ctx context.Context, cmd *command.RollbackAgentCommand) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".RollbackAgent")
	defer span.End()

	agentID, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}

	agent, err := svc.domainService.Rollback(spanCtx, agentID, cmd.ReleaseID)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) GetRelease(ctx context.Context, releaseID string) (*response.AgentReleaseResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetRelease")
	defer span.End()

	release, err := svc.domainService.GetRelease(spanCtx, releaseID)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toReleaseResponse(release), nil
}

func (svc *agentApplicationService) ListReleases(ctx context.Context, agentID string, page, pageSize int) (*response.AgentReleaseListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListReleases")
	defer span.End()

	id, err := uuid.Parse(agentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}

	releases, total, err := svc.domainService.ListReleases(spanCtx, id, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := make([]response.AgentReleaseResponse, 0, len(releases))
	for _, r := range releases {
		records = append(records, *toReleaseResponse(&r))
	}
	return &response.AgentReleaseListResponse{Record: records, Total: total}, nil
}

// --- 转换辅助函数 ---

func toAgentResponse(a *agentModel.Agent) *response.AgentResponse {
	return &response.AgentResponse{
		AgentID:      a.ID.String(),
		WorkspaceID:  a.WorkspaceID.String(),
		Name:         a.Name,
		Description:  a.Description,
		AgentType:    a.AgentType.String(),
		SystemPrompt: a.SystemPrompt,
		ModelConfig: response.ModelConfigResponse{
			ModelInstanceID: a.ModelConfig.ModelInstanceID,
			Parameters:      a.ModelConfig.Parameters,
		},
		ToolBindings:      a.ToolBindings,
		KnowledgeBindings: a.KnowledgeBindings,
		SkillBindings:     a.SkillBindings,
		MCPBindings:       a.MCPBindings,
		MemoryConfig: response.MemoryConfigResponse{
			MemoryDriverType:  a.MemoryConfig.MemoryDriverType,
			ContextMode:       a.MemoryConfig.ContextMode,
			ContextWindowSize: a.MemoryConfig.ContextWindowSize,
			SummarizationConfig: response.SummarizationConfigResponse{
				SummaryModelInstanceID: a.MemoryConfig.SummarizationConfig.SummaryModelInstanceID,
				TriggerTokenThreshold:  a.MemoryConfig.SummarizationConfig.TriggerTokenThreshold,
			},
			GlobalMemoryMode: a.MemoryConfig.GlobalMemoryMode,
		},
		CollaborationConfig: response.CollaborationConfigResponse{
			SubAgentIDs:        a.CollaborationConfig.SubAgentIDs,
			TransferPolicy:     a.CollaborationConfig.TransferPolicy,
			MaxDelegationDepth: a.CollaborationConfig.MaxDelegationDepth,
		},
		SandboxConfig: response.SandboxConfigResponse{
			Enabled:     a.SandboxConfig.Enabled,
			SandboxType: a.SandboxConfig.SandboxType,
		},
		ActiveReleaseID: a.ActiveReleaseID,
		Status:          a.Status.String(),
		CreatedAt:       a.CreatedAt,
	}
}

func toReleaseResponse(r *agentModel.AgentRelease) *response.AgentReleaseResponse {
	return &response.AgentReleaseResponse{
		ReleaseID:  r.ReleaseID,
		AgentID:    r.AgentID.String(),
		ChangeLog:  r.ChangeLog,
		Status:     r.Status.String(),
		ReleasedAt: r.ReleasedAt,
		ReleasedBy: r.ReleasedBy.String(),
	}
}
