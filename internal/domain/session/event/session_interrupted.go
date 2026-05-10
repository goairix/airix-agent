package event

import (
	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/shared/event"
)

const TypeSessionInterrupted = "session.interrupted"

type SessionInterruptedPayload struct {
	SessionID      uuid.UUID
	AgentID        uuid.UUID
	WorkspaceID    uuid.UUID
	InterruptID    string
	PendingContext string
}

func NewSessionInterrupted(sessionID, agentID, workspaceID uuid.UUID, interruptID, pendingContext string) event.DomainEvent[SessionInterruptedPayload] {
	return event.NewDomainEvent(
		TypeSessionInterrupted,
		sessionID.String(),
		"Session",
		SessionInterruptedPayload{
			SessionID:      sessionID,
			AgentID:        agentID,
			WorkspaceID:    workspaceID,
			InterruptID:    interruptID,
			PendingContext: pendingContext,
		},
	)
}
