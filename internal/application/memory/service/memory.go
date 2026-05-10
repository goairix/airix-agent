package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/application/memory/dto/response"
	memoryModel "github.com/dysodeng/app/internal/domain/memory/model"
	"github.com/dysodeng/app/internal/domain/memory/repository"
	"github.com/dysodeng/app/internal/domain/memory/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Memory 应用服务接口
type Service interface {
	SaveMemory(ctx context.Context, workspaceID, userID, agentID, content string, memoryType uint8, tags []string, importance float64) (*response.MemoryResponse, error)
	ListByDate(ctx context.Context, workspaceID, userID string, date time.Time) ([]response.MemoryResponse, error)
}

type memoryApplicationService struct {
	baseTraceSpanName string
	memoryRepo        repository.MemoryRepository
}

func NewMemoryApplicationService(memoryRepo repository.MemoryRepository) Service {
	return &memoryApplicationService{
		baseTraceSpanName: "application.memory.service.MemoryApplicationService",
		memoryRepo:        memoryRepo,
	}
}

func (svc *memoryApplicationService) SaveMemory(ctx context.Context, workspaceID, userID, agentID, content string, memoryType uint8, tags []string, importance float64) (*response.MemoryResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".SaveMemory")
	defer span.End()

	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("用户 ID 格式错误")
	}

	mt := valueobject.MemoryType(memoryType)
	m := memoryModel.NewMemory(wid, uid, mt, content, tags, importance)
	if agentID != "" {
		aid, err := uuid.Parse(agentID)
		if err != nil {
			return nil, errors.New("Agent ID 格式错误")
		}
		m.AgentID = aid
	}
	if err = svc.memoryRepo.Save(spanCtx, m); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toMemoryResponse(m), nil
}

func (svc *memoryApplicationService) ListByDate(ctx context.Context, workspaceID, userID string, date time.Time) ([]response.MemoryResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListByDate")
	defer span.End()

	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("用户 ID 格式错误")
	}

	memories, err := svc.memoryRepo.ListByUserAndDate(spanCtx, wid, uid, date)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	result := make([]response.MemoryResponse, 0, len(memories))
	for _, m := range memories {
		result = append(result, *toMemoryResponse(&m))
	}
	return result, nil
}

func toMemoryResponse(m *memoryModel.Memory) *response.MemoryResponse {
	return &response.MemoryResponse{
		MemoryID:    m.ID.String(),
		WorkspaceID: m.WorkspaceID.String(),
		UserID:      m.UserID.String(),
		AgentID:     m.AgentID.String(),
		MemoryType:  m.MemoryType.String(),
		Content:     m.Content,
		Tags:        m.Tags,
		Importance:  m.Importance,
		Date:        m.Date,
		CreatedAt:   m.CreatedAt,
	}
}
