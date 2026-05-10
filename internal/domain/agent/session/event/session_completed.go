package event

import (
	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/shared/event"
)

const TypeSessionCompleted = "session.completed"

type SessionCompletedPayload struct {
	SessionID   uuid.UUID
	AgentID     uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

func NewSessionCompleted(sessionID, agentID, workspaceID, userID uuid.UUID) event.DomainEvent[SessionCompletedPayload] {
	return event.NewDomainEvent(
		TypeSessionCompleted,
		sessionID.String(),
		"Session",
		SessionCompletedPayload{
			SessionID:   sessionID,
			AgentID:     agentID,
			WorkspaceID: workspaceID,
			UserID:      userID,
		},
	)
}
