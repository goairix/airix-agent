package model

import (
	"time"

	"github.com/google/uuid"

	sessionErrors "github.com/dysodeng/app/internal/domain/session/errors"
	"github.com/dysodeng/app/internal/domain/session/valueobject"
)

// TokenUsage Token 消耗
type TokenUsage struct {
	InputTokens  int64
	OutputTokens int64
	CachedTokens int64
}

// InterruptState 中断状态（Status = interrupted 时有值）
type InterruptState struct {
	InterruptID    string // ADK StatefulInterrupt 唯一标识
	CheckPointData string // 序列化的 CheckPointStore 快照
	PendingContext string // 等待人工输入的上下文描述
}

// Session 聚合根
type Session struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	AgentID         uuid.UUID
	ReleaseID       string
	UserID          uuid.UUID
	Title           string
	Status          valueobject.SessionStatus
	TotalTokenUsage TokenUsage
	InterruptState  *InterruptState
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
}

func NewSession(workspaceID, agentID, userID uuid.UUID, releaseID, title string) (*Session, error) {
	id, _ := uuid.NewV7()
	s := &Session{
		ID:          id,
		WorkspaceID: workspaceID,
		AgentID:     agentID,
		UserID:      userID,
		ReleaseID:   releaseID,
		Title:       title,
		Status:      valueobject.SessionStatusRunning,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) Validate() error {
	if s.AgentID == uuid.Nil {
		return sessionErrors.ErrSessionAgentIDEmpty
	}
	if s.WorkspaceID == uuid.Nil {
		return sessionErrors.ErrSessionWorkspaceEmpty
	}
	return nil
}

func (s *Session) Complete() { s.Status = valueobject.SessionStatusCompleted }
func (s *Session) Fail()     { s.Status = valueobject.SessionStatusFailed }
func (s *Session) Interrupt(state *InterruptState) {
	s.Status = valueobject.SessionStatusInterrupted
	s.InterruptState = state
}
