package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/application/agent/session/dto/command"
	"github.com/dysodeng/app/internal/application/agent/session/dto/response"
	sessionErrors "github.com/dysodeng/app/internal/domain/agent/session/errors"
	sessionModel "github.com/dysodeng/app/internal/domain/agent/session/model"
	"github.com/dysodeng/app/internal/domain/agent/session/repository"
	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Session 应用服务接口
type Service interface {
	CreateSession(ctx context.Context, cmd *command.CreateSessionCommand) (*response.SessionResponse, error)
	GetSession(ctx context.Context, sessionID string) (*response.SessionResponse, error)
	CreateMessage(ctx context.Context, sessionID, query string, agentInput map[string]any) (*response.MessageResponse, error)
	CompleteMessage(ctx context.Context, cmd *command.CompleteMessageCommand) error
}

type sessionApplicationService struct {
	baseTraceSpanName string
	sessionRepo       repository.SessionRepository
	messageRepo       repository.MessageRepository
}

func NewSessionApplicationService(
	sessionRepo repository.SessionRepository,
	messageRepo repository.MessageRepository,
) Service {
	return &sessionApplicationService{
		baseTraceSpanName: "application.session.service.SessionApplicationService",
		sessionRepo:       sessionRepo,
		messageRepo:       messageRepo,
	}
}

func (svc *sessionApplicationService) CreateSession(ctx context.Context, cmd *command.CreateSessionCommand) (*response.SessionResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateSession")
	defer span.End()

	workspaceID, err := uuid.Parse(cmd.WorkspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	agentID, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, errors.New("用户 ID 格式错误")
	}

	session, err := sessionModel.NewSession(workspaceID, agentID, userID, cmd.ReleaseID, cmd.Title)
	if err != nil {
		return nil, err
	}
	if err = svc.sessionRepo.Save(spanCtx, session); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toSessionResponse(session), nil
}

func (svc *sessionApplicationService) GetSession(ctx context.Context, sessionID string) (*response.SessionResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetSession")
	defer span.End()

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("会话 ID 格式错误")
	}
	session, err := svc.sessionRepo.FindByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	if session == nil {
		return nil, sessionErrors.ErrSessionNotFound
	}
	return toSessionResponse(session), nil
}

func (svc *sessionApplicationService) CreateMessage(ctx context.Context, sessionID, query string, agentInput map[string]any) (*response.MessageResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateMessage")
	defer span.End()

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("会话 ID 格式错误")
	}
	session, err := svc.sessionRepo.FindByID(spanCtx, sid)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	if session == nil {
		return nil, sessionErrors.ErrSessionNotFound
	}
	msg := sessionModel.NewMessage(session.ID, session.WorkspaceID, session.AgentID, query, time.Now().UnixMilli())
	msg.AgentInput = agentInput
	if err = svc.messageRepo.Save(spanCtx, msg); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toMessageResponse(msg), nil
}

func (svc *sessionApplicationService) CompleteMessage(ctx context.Context, cmd *command.CompleteMessageCommand) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CompleteMessage")
	defer span.End()

	msgID, err := uuid.Parse(cmd.MessageID)
	if err != nil {
		return errors.New("消息 ID 格式错误")
	}
	msg, err := svc.messageRepo.FindByID(spanCtx, msgID)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	if msg == nil {
		return sessionErrors.ErrMessageQueryFailed
	}
	now := time.Now()
	msg.Status = valueobject.MessageStatusCompleted
	msg.TotalTokens = cmd.TotalTokens
	msg.InputTokens = cmd.InputTokens
	msg.OutputTokens = cmd.OutputTokens
	msg.CachedTokens = cmd.CachedTokens
	msg.ExecutionTimeMs = cmd.ExecutionTimeMs
	msg.FirstTokenLatencyMs = cmd.FirstTokenLatencyMs
	msg.CompletedAt = &now
	if err = svc.messageRepo.Update(spanCtx, msg); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func toSessionResponse(s *sessionModel.Session) *response.SessionResponse {
	return &response.SessionResponse{
		SessionID:   s.ID.String(),
		WorkspaceID: s.WorkspaceID.String(),
		AgentID:     s.AgentID.String(),
		ReleaseID:   s.ReleaseID,
		UserID:      s.UserID.String(),
		Title:       s.Title,
		Status:      s.Status.String(),
		CreatedAt:   s.CreatedAt,
	}
}

func toMessageResponse(m *sessionModel.Message) *response.MessageResponse {
	return &response.MessageResponse{
		MessageID:   m.ID.String(),
		SessionID:   m.SessionID.String(),
		Query:       m.Query,
		Status:      m.Status.String(),
		TotalTokens: m.TotalTokens,
		CreatedAt:   m.CreatedAt,
	}
}
