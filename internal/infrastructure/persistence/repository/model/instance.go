package model

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	modelEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/model"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	"github.com/dysodeng/app/internal/infrastructure/pkg/crypto/aes"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// InstanceRepositoryConfig 仓储加密配置
type InstanceRepositoryConfig struct {
	EncryptKey []byte
	EncryptIV  []byte
}

type instanceRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
	encryptKey        []byte
	encryptIV         []byte
}

func NewInstanceRepository(txManager transactions.TransactionManager, cfg InstanceRepositoryConfig) repository.InstanceRepository {
	return &instanceRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.model.InstanceRepository",
		txManager:         txManager,
		encryptKey:        cfg.EncryptKey,
		encryptIV:         cfg.EncryptIV,
	}
}

func (repo *instanceRepository) Save(ctx context.Context, inst *domainModel.Instance) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(inst)
	if err != nil {
		return err
	}

	var exists modelEntity.Instance
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

func (repo *instanceRepository) FindByID(ctx context.Context, instanceID uuid.UUID) (*domainModel.Instance, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity modelEntity.Instance
	if err := tx.Where("id = ?", instanceID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *instanceRepository) FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]domainModel.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByWorkspace")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []modelEntity.Instance
	var total int64

	query := tx.Model(&modelEntity.Instance{}).Where("workspace_id = ?", workspaceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domainModel.Instance, 0, len(entities))
	for _, e := range entities {
		inst, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *inst)
	}
	return result, total, nil
}

func (repo *instanceRepository) FindByWorkspaceAndCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination repository.Pagination) ([]domainModel.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByWorkspaceAndCapability")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []modelEntity.Instance
	var total int64

	query := tx.Model(&modelEntity.Instance{}).Where("workspace_id = ? AND capability = ?", workspaceID, capability.Uint8())
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domainModel.Instance, 0, len(entities))
	for _, e := range entities {
		inst, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *inst)
	}
	return result, total, nil
}

func (repo *instanceRepository) Delete(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Delete")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("id = ?", instanceID).Delete(&modelEntity.Instance{}).Error
}

func (repo *instanceRepository) ExistsByProviderID(ctx context.Context, providerID uuid.UUID) (bool, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ExistsByProviderID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var count int64
	if err := tx.Model(&modelEntity.Instance{}).Where("provider_id = ?", providerID).Limit(1).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *instanceRepository) encryptAPIKey(plainKey string) (string, error) {
	if plainKey == "" {
		return "", nil
	}
	encrypted, err := aes.Encrypt([]byte(plainKey), repo.encryptKey, repo.encryptIV)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (repo *instanceRepository) decryptAPIKey(encryptedKey string) (string, error) {
	if encryptedKey == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}
	decrypted, err := aes.Decrypt(data, repo.encryptKey, repo.encryptIV)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func (repo *instanceRepository) fromEntity(e *modelEntity.Instance) (*domainModel.Instance, error) {
	apiKey, err := repo.decryptAPIKey(e.APIKey)
	if err != nil {
		return nil, err
	}
	var parameters map[string]any
	if e.Parameters != "" {
		if err = json.Unmarshal([]byte(e.Parameters), &parameters); err != nil {
			return nil, err
		}
	}
	return &domainModel.Instance{
		ID:          e.ID,
		WorkspaceID: e.WorkspaceID,
		ProviderID:  e.ProviderID,
		ModelName:   e.ModelName,
		Capability:  valueobject.Capability(e.Capability),
		APIKey:      apiKey,
		Parameters:  parameters,
		RateLimit:   domainModel.RateLimit{RPM: e.RateLimitRPM, TPM: e.RateLimitTPM},
		Status:      valueobject.InstanceStatus(e.Status),
		CreatedAt:   e.CreatedAt.Time,
		UpdatedAt:   e.UpdatedAt.Time,
	}, nil
}

func (repo *instanceRepository) toEntity(inst *domainModel.Instance) (*modelEntity.Instance, error) {
	encryptedKey, err := repo.encryptAPIKey(inst.APIKey)
	if err != nil {
		return nil, err
	}
	paramsJSON, err := json.Marshal(inst.Parameters)
	if err != nil {
		return nil, err
	}
	return &modelEntity.Instance{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: inst.ID},
		WorkspaceID:             inst.WorkspaceID,
		ProviderID:              inst.ProviderID,
		ModelName:               inst.ModelName,
		Capability:              inst.Capability.Uint8(),
		APIKey:                  encryptedKey,
		Parameters:              string(paramsJSON),
		RateLimitRPM:            inst.RateLimit.RPM,
		RateLimitTPM:            inst.RateLimit.TPM,
		Status:                  inst.Status.Uint8(),
	}, nil
}
