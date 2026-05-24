package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/example/go-diamond/internal/model"
	"github.com/example/go-diamond/internal/store"
)

type ConfigService struct {
	configStore  *store.ConfigStore
	historyStore *store.HistoryStore
}

func NewConfigService(configStore *store.ConfigStore, historyStore *store.HistoryStore) *ConfigService {
	return &ConfigService{
		configStore:  configStore,
		historyStore: historyStore,
	}
}

func (s *ConfigService) GetConfig(ctx context.Context, namespace, group, dataID string) (*model.Config, error) {
	return s.configStore.Get(ctx, namespace, group, dataID)
}

func (s *ConfigService) CreateConfig(ctx context.Context, cfg *model.Config) error {
	cfg.ContentMD5 = md5Hash(cfg.Content)
	cfg.Version = 1

	if err := s.configStore.Create(ctx, cfg); err != nil {
		return err
	}

	history := &model.ConfigHistory{
		ConfigID:   cfg.ID,
		Namespace:  cfg.Namespace,
		Group:      cfg.Group,
		DataID:     cfg.DataID,
		Content:    cfg.Content,
		ContentMD5: cfg.ContentMD5,
		Version:    cfg.Version,
		OpType:     "create",
		OpBy:       cfg.CreatedBy,
	}

	return s.historyStore.Create(ctx, history)
}

func (s *ConfigService) UpdateConfig(ctx context.Context, cfg *model.Config) error {
	existing, err := s.configStore.Get(ctx, cfg.Namespace, cfg.Group, cfg.DataID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("config not found")
	}

	cfg.ID = existing.ID
	cfg.ContentMD5 = md5Hash(cfg.Content)
	cfg.Version = existing.Version + 1

	if err := s.configStore.Update(ctx, cfg); err != nil {
		return err
	}

	history := &model.ConfigHistory{
		ConfigID:   existing.ID,
		Namespace:  cfg.Namespace,
		Group:      cfg.Group,
		DataID:     cfg.DataID,
		Content:    cfg.Content,
		ContentMD5: cfg.ContentMD5,
		Version:    cfg.Version,
		OpType:     "update",
		OpBy:       cfg.UpdatedBy,
	}

	return s.historyStore.Create(ctx, history)
}

func (s *ConfigService) DeleteConfig(ctx context.Context, namespace, group, dataID, operator string) error {
	existing, err := s.configStore.Get(ctx, namespace, group, dataID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("config not found")
	}

	if err := s.configStore.Delete(ctx, namespace, group, dataID, operator); err != nil {
		return err
	}

	history := &model.ConfigHistory{
		ConfigID:   existing.ID,
		Namespace:  namespace,
		Group:      group,
		DataID:     dataID,
		Content:    existing.Content,
		ContentMD5: existing.ContentMD5,
		Version:    existing.Version,
		OpType:     "delete",
		OpBy:       operator,
	}

	return s.historyStore.Create(ctx, history)
}

func (s *ConfigService) ListConfigs(ctx context.Context, namespace, group string, page, pageSize int) ([]model.Config, int, error) {
	return s.configStore.List(ctx, namespace, group, page, pageSize)
}

func (s *ConfigService) GetHistories(ctx context.Context, configID uint64, page, pageSize int) ([]model.ConfigHistory, int, error) {
	return s.historyStore.List(ctx, configID, page, pageSize)
}

func (s *ConfigService) Rollback(ctx context.Context, namespace, group, dataID string, historyID uint64, operator string) error {
	existing, err := s.configStore.Get(ctx, namespace, group, dataID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("config not found")
	}

	history, err := s.historyStore.GetByID(ctx, historyID)
	if err != nil {
		return err
	}
	if history == nil {
		return fmt.Errorf("history not found")
	}

	cfg := &model.Config{
		Namespace:   namespace,
		Group:       group,
		DataID:      dataID,
		Content:     history.Content,
		Format:      existing.Format,
		Description: existing.Description,
		UpdatedBy:   operator,
	}

	return s.UpdateConfig(ctx, cfg)
}

func md5Hash(content string) string {
	h := md5.Sum([]byte(content))
	return hex.EncodeToString(h[:])
}