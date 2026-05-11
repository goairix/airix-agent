// internal/domain/model/model/instance_test.go
package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

func TestNewInstance(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	inst, err := NewInstance(wsID, providerID, "gpt-4o", valueobject.CapabilityChat)
	assert.NoError(t, err)
	assert.NotEmpty(t, inst.ID)
	assert.Equal(t, wsID, inst.WorkspaceID)
	assert.Equal(t, providerID, inst.ProviderID)
	assert.Equal(t, "gpt-4o", inst.ModelName)
	assert.Equal(t, valueobject.CapabilityChat, inst.Capability)
	assert.Equal(t, valueobject.InstanceStatusActive, inst.Status)
}

func TestNewInstance_EmptyWorkspace(t *testing.T) {
	providerID, _ := uuid.NewV7()
	_, err := NewInstance(uuid.Nil, providerID, "gpt-4o", valueobject.CapabilityChat)
	assert.Error(t, err)
}

func TestNewInstance_EmptyModelName(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	_, err := NewInstance(wsID, providerID, "", valueobject.CapabilityChat)
	assert.Error(t, err)
}

func TestInstance_DisableAndEnable(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	inst, _ := NewInstance(wsID, providerID, "gpt-4o", valueobject.CapabilityChat)
	inst.Disable()
	assert.Equal(t, valueobject.InstanceStatusDisabled, inst.Status)
	inst.Enable()
	assert.Equal(t, valueobject.InstanceStatusActive, inst.Status)
}

func TestInstance_SetAPIKey(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	inst, _ := NewInstance(wsID, providerID, "gpt-4o", valueobject.CapabilityChat)
	inst.SetAPIKey("sk-test-key-123")
	assert.Equal(t, "sk-test-key-123", inst.APIKey)
}
