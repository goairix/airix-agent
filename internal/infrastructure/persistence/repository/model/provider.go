package model

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	modelEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/model"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type providerRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewProviderRepository(txManager transactions.TransactionManager) repository.ProviderRepository {
	return &providerRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.model.ProviderRepository",
		txManager:         txManager,
	}
}

func (repo *providerRepository) Save(ctx context.Context, p *domainModel.Provider) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(p)
	if err != nil {
		return err
	}

	var exists modelEntity.Provider
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
	return nil
}

func (repo *providerRepository) FindByID(ctx context.Context, providerID uuid.UUID) (*domainModel.Provider, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity modelEntity.Provider
	if err := tx.Where("id = ?", providerID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *providerRepository) FindAll(ctx context.Context, pagination repository.Pagination) ([]domainModel.Provider, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindAll")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []modelEntity.Provider
	var total int64

	query := tx.Model(&modelEntity.Provider{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domainModel.Provider, 0, len(entities))
	for _, e := range entities {
		p, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *p)
	}
	return result, total, nil
}

func (repo *providerRepository) Delete(ctx context.Context, providerID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Delete")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("id = ?", providerID).Delete(&modelEntity.Provider{}).Error
}

func (repo *providerRepository) fromEntity(e *modelEntity.Provider) (*domainModel.Provider, error) {
	var caps []valueobject.Capability
	if e.SupportedCapabilities != "" {
		var rawCaps []uint8
		if err := json.Unmarshal([]byte(e.SupportedCapabilities), &rawCaps); err != nil {
			return nil, err
		}
		for _, c := range rawCaps {
			caps = append(caps, valueobject.Capability(c))
		}
	}
	return &domainModel.Provider{
		ID:                    e.ID,
		Name:                  e.Name,
		Protocol:              valueobject.Protocol(e.Protocol),
		BaseURL:               e.BaseURL,
		AuthType:              valueobject.AuthType(e.AuthType),
		SupportedCapabilities: caps,
		CreatedAt:             e.CreatedAt.Time,
		UpdatedAt:             e.UpdatedAt.Time,
	}, nil
}

func (repo *providerRepository) toEntity(p *domainModel.Provider) (*modelEntity.Provider, error) {
	rawCaps := make([]uint8, 0, len(p.SupportedCapabilities))
	for _, c := range p.SupportedCapabilities {
		rawCaps = append(rawCaps, c.Uint8())
	}
	capsJSON, err := json.Marshal(rawCaps)
	if err != nil {
		return nil, err
	}
	return &modelEntity.Provider{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: p.ID},
		Name:                    p.Name,
		Protocol:                p.Protocol.Uint8(),
		BaseURL:                 p.BaseURL,
		AuthType:                p.AuthType.Uint8(),
		SupportedCapabilities:   string(capsJSON),
	}, nil
}
