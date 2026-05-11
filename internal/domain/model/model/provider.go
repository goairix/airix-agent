// internal/domain/model/model/provider.go
package model

import (
	"time"

	"github.com/google/uuid"

	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// Provider 模型厂商聚合根（系统级，无 WorkspaceID）
type Provider struct {
	ID                    uuid.UUID
	Name                  string
	Protocol              valueobject.Protocol
	BaseURL               string
	AuthType              valueobject.AuthType
	SupportedCapabilities []valueobject.Capability
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func NewProvider(name string, protocol valueobject.Protocol, baseURL string, authType valueobject.AuthType) (*Provider, error) {
	id, _ := uuid.NewV7()
	p := &Provider{
		ID:       id,
		Name:     name,
		Protocol: protocol,
		BaseURL:  baseURL,
		AuthType: authType,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Provider) Validate() error {
	if p.Name == "" {
		return modelErrors.ErrProviderNameEmpty
	}
	if err := p.Protocol.Validate(); err != nil {
		return modelErrors.ErrProviderProtocolInvalid
	}
	if err := p.AuthType.Validate(); err != nil {
		return modelErrors.ErrProviderAuthTypeInvalid
	}
	return nil
}

// AddCapability 添加支持的能力（去重）
func (p *Provider) AddCapability(cap valueobject.Capability) {
	for _, c := range p.SupportedCapabilities {
		if c == cap {
			return
		}
	}
	p.SupportedCapabilities = append(p.SupportedCapabilities, cap)
}

// SupportsCapability 检查是否支持指定能力
func (p *Provider) SupportsCapability(cap valueobject.Capability) bool {
	for _, c := range p.SupportedCapabilities {
		if c == cap {
			return true
		}
	}
	return false
}
