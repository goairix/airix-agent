package model

import (
	"time"

	"github.com/google/uuid"

	wsErrors "github.com/dysodeng/app/internal/domain/workspace/errors"
	"github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

// Workspace 工作空间聚合根
type Workspace struct {
	ID          uuid.UUID
	Name        string
	Description string
	Status      valueobject.Status
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewWorkspace(name, description string, createdBy uuid.UUID) (*Workspace, error) {
	id, _ := uuid.NewV7()
	w := &Workspace{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      valueobject.StatusActive,
		CreatedBy:   createdBy,
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Workspace) Validate() error {
	if w.Name == "" {
		return wsErrors.ErrWorkspaceNameEmpty
	}
	return nil
}

func (w *Workspace) Disable() {
	w.Status = valueobject.StatusDisabled
}

func (w *Workspace) Enable() {
	w.Status = valueobject.StatusActive
}
