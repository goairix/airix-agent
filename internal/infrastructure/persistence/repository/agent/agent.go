package agent

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	agentEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type agentRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewAgentRepository(txManager transactions.TransactionManager) repository.Repository {
	return &agentRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.agent.Repository",
		txManager:         txManager,
	}
}

func (repo *agentRepository) Save(ctx context.Context, a *agentModel.Agent) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(a)
	if err != nil {
		return err
	}

	if a.ID != uuid.Nil {
		var exists agentEntity.Agent
		tx.Where("id = ?", entity.ID).First(&exists)
		if exists.ID == uuid.Nil {
			if err = tx.Create(entity).Error; err != nil {
				return err
			}
		} else {
			if err = tx.Where("id = ?", entity.ID).Updates(entity).Error; err != nil {
				return err
			}
		}
	} else {
		if err = tx.Create(entity).Error; err != nil {
			return err
		}
		a.ID = entity.ID
		a.CreatedAt = entity.CreatedAt.Time
	}
	return nil
}

func (repo *agentRepository) FindByID(ctx context.Context, agentID uuid.UUID) (*agentModel.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity agentEntity.Agent
	if err := tx.Where("id = ?", agentID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *agentRepository) FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]agentModel.Agent, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByWorkspace")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []agentEntity.Agent
	var total int64

	query := tx.Model(&agentEntity.Agent{}).Where("workspace_id = ?", workspaceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]agentModel.Agent, 0, len(entities))
	for _, e := range entities {
		a, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *a)
	}
	return result, total, nil
}

func (repo *agentRepository) Delete(ctx context.Context, agentID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Delete")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("id = ?", agentID).Delete(&agentEntity.Agent{}).Error
}

// --- 转换方法 ---

func (repo *agentRepository) fromEntity(e *agentEntity.Agent) (*agentModel.Agent, error) {
	var modelConfig agentModel.ModelConfig
	if e.ModelConfig != "" {
		if err := json.Unmarshal([]byte(e.ModelConfig), &modelConfig); err != nil {
			return nil, err
		}
	}
	var memoryConfig agentModel.MemoryConfig
	if e.MemoryConfig != "" {
		if err := json.Unmarshal([]byte(e.MemoryConfig), &memoryConfig); err != nil {
			return nil, err
		}
	}
	var collaborationConfig agentModel.CollaborationConfig
	if e.CollaborationConfig != "" {
		if err := json.Unmarshal([]byte(e.CollaborationConfig), &collaborationConfig); err != nil {
			return nil, err
		}
	}
	var sandboxConfig agentModel.SandboxConfig
	if e.SandboxConfig != "" {
		if err := json.Unmarshal([]byte(e.SandboxConfig), &sandboxConfig); err != nil {
			return nil, err
		}
	}
	var toolBindings []string
	if e.ToolBindings != "" {
		if err := json.Unmarshal([]byte(e.ToolBindings), &toolBindings); err != nil {
			return nil, err
		}
	}
	var knowledgeBindings []string
	if e.KnowledgeBindings != "" {
		if err := json.Unmarshal([]byte(e.KnowledgeBindings), &knowledgeBindings); err != nil {
			return nil, err
		}
	}
	var skillBindings []string
	if e.SkillBindings != "" {
		if err := json.Unmarshal([]byte(e.SkillBindings), &skillBindings); err != nil {
			return nil, err
		}
	}
	var mcpBindings []string
	if e.MCPBindings != "" {
		if err := json.Unmarshal([]byte(e.MCPBindings), &mcpBindings); err != nil {
			return nil, err
		}
	}

	return &agentModel.Agent{
		ID:                  e.ID,
		WorkspaceID:         e.WorkspaceID,
		Name:                e.Name,
		Description:         e.Description,
		AgentType:           valueobject.AgentType(e.AgentType),
		SystemPrompt:        e.SystemPrompt,
		ModelConfig:         modelConfig,
		ToolBindings:        toolBindings,
		KnowledgeBindings:   knowledgeBindings,
		SkillBindings:       skillBindings,
		MCPBindings:         mcpBindings,
		MemoryConfig:        memoryConfig,
		CollaborationConfig: collaborationConfig,
		SandboxConfig:       sandboxConfig,
		ActiveReleaseID:     e.ActiveReleaseID,
		Status:              valueobject.AgentStatus(e.Status),
		CreatedBy:           e.CreatedBy,
		CreatedAt:           e.CreatedAt.Time,
		UpdatedAt:           e.UpdatedAt.Time,
	}, nil
}

func (repo *agentRepository) toEntity(a *agentModel.Agent) (*agentEntity.Agent, error) {
	modelConfigJSON, err := json.Marshal(a.ModelConfig)
	if err != nil {
		return nil, err
	}
	memoryConfigJSON, err := json.Marshal(a.MemoryConfig)
	if err != nil {
		return nil, err
	}
	collaborationConfigJSON, err := json.Marshal(a.CollaborationConfig)
	if err != nil {
		return nil, err
	}
	sandboxConfigJSON, err := json.Marshal(a.SandboxConfig)
	if err != nil {
		return nil, err
	}
	toolBindingsJSON, err := json.Marshal(a.ToolBindings)
	if err != nil {
		return nil, err
	}
	knowledgeBindingsJSON, err := json.Marshal(a.KnowledgeBindings)
	if err != nil {
		return nil, err
	}
	skillBindingsJSON, err := json.Marshal(a.SkillBindings)
	if err != nil {
		return nil, err
	}
	mcpBindingsJSON, err := json.Marshal(a.MCPBindings)
	if err != nil {
		return nil, err
	}

	return &agentEntity.Agent{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: a.ID},
		WorkspaceID:             a.WorkspaceID,
		Name:                    a.Name,
		Description:             a.Description,
		AgentType:               a.AgentType.Uint8(),
		SystemPrompt:            a.SystemPrompt,
		ModelConfig:             string(modelConfigJSON),
		ToolBindings:            string(toolBindingsJSON),
		KnowledgeBindings:       string(knowledgeBindingsJSON),
		SkillBindings:           string(skillBindingsJSON),
		MCPBindings:             string(mcpBindingsJSON),
		MemoryConfig:            string(memoryConfigJSON),
		CollaborationConfig:     string(collaborationConfigJSON),
		SandboxConfig:           string(sandboxConfigJSON),
		ActiveReleaseID:         a.ActiveReleaseID,
		Status:                  a.Status.Uint8(),
		CreatedBy:               a.CreatedBy,
	}, nil
}
