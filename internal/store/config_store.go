package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/go-diamond/internal/model"
	"github.com/jmoiron/sqlx"
)

type ConfigStore struct {
	db *sqlx.DB
}

func NewConfigStore(db *sqlx.DB) *ConfigStore {
	return &ConfigStore{db: db}
}

func (s *ConfigStore) Get(ctx context.Context, namespace, group, dataID string) (*model.Config, error) {
	var cfg model.Config
	query := `SELECT id, namespace, ` + "`group`" + `, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at
			  FROM configs WHERE namespace=? AND ` + "`group`" + `=? AND data_id=? AND is_deleted=0`
	err := s.db.GetContext(ctx, &cfg, query, namespace, group, dataID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *ConfigStore) Create(ctx context.Context, cfg *model.Config) error {
	query := `INSERT INTO configs (namespace, ` + "`group`" + `, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, NOW(), NOW())`
	result, err := s.db.ExecContext(ctx, query, cfg.Namespace, cfg.Group, cfg.DataID, cfg.Content, cfg.ContentMD5, cfg.Format, cfg.Description, cfg.Version, cfg.CreatedBy, cfg.UpdatedBy)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	cfg.ID = uint64(id)
	return nil
}

func (s *ConfigStore) Update(ctx context.Context, cfg *model.Config) error {
	query := `UPDATE configs SET content=?, content_md5=?, format=?, description=?, version=?, updated_by=?, updated_at=NOW()
			  WHERE namespace=? AND ` + "`group`" + `=? AND data_id=? AND is_deleted=0`
	_, err := s.db.ExecContext(ctx, query, cfg.Content, cfg.ContentMD5, cfg.Format, cfg.Description, cfg.Version, cfg.UpdatedBy, cfg.Namespace, cfg.Group, cfg.DataID)
	return err
}

func (s *ConfigStore) Delete(ctx context.Context, namespace, group, dataID, operator string) error {
	query := `UPDATE configs SET is_deleted=1, updated_by=?, updated_at=NOW() WHERE namespace=? AND ` + "`group`" + `=? AND data_id=? AND is_deleted=0`
	_, err := s.db.ExecContext(ctx, query, operator, namespace, group, dataID)
	return err
}

func (s *ConfigStore) List(ctx context.Context, namespace, group string, page, pageSize int) ([]model.Config, int, error) {
	offset := (page - 1) * pageSize
	var configs []model.Config
	var total int

	countQuery := `SELECT COUNT(*) FROM configs WHERE namespace=? AND ` + "`group`" + `=? AND is_deleted=0`
	err := s.db.GetContext(ctx, &total, countQuery, namespace, group)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, namespace, ` + "`group`" + `, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at
			  FROM configs WHERE namespace=? AND ` + "`group`" + `=? AND is_deleted=0 ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	err = s.db.SelectContext(ctx, &configs, query, namespace, group, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return configs, total, nil
}

func (s *ConfigStore) GetUpdatedConfigs(ctx context.Context, since time.Time) ([]model.Config, error) {
	var configs []model.Config
	query := `SELECT id, namespace, ` + "`group`" + `, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at
			  FROM configs WHERE updated_at > ? AND is_deleted=0`
	err := s.db.SelectContext(ctx, &configs, query, since)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

type HistoryStore struct {
	db *sqlx.DB
}

func NewHistoryStore(db *sqlx.DB) *HistoryStore {
	return &HistoryStore{db: db}
}

func (s *HistoryStore) Create(ctx context.Context, h *model.ConfigHistory) error {
	query := `INSERT INTO config_histories (config_id, namespace, ` + "`group`" + `, data_id, content, content_md5, version, op_type, op_by, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`
	_, err := s.db.ExecContext(ctx, query, h.ConfigID, h.Namespace, h.Group, h.DataID, h.Content, h.ContentMD5, h.Version, h.OpType, h.OpBy)
	return err
}

func (s *HistoryStore) List(ctx context.Context, configID uint64, page, pageSize int) ([]model.ConfigHistory, int, error) {
	offset := (page - 1) * pageSize
	var histories []model.ConfigHistory
	var total int

	countQuery := `SELECT COUNT(*) FROM config_histories WHERE config_id=?`
	err := s.db.GetContext(ctx, &total, countQuery, configID)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, config_id, namespace, ` + "`group`" + `, data_id, content, content_md5, version, op_type, op_by, created_at
			  FROM config_histories WHERE config_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err = s.db.SelectContext(ctx, &histories, query, configID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return histories, total, nil
}

func (s *HistoryStore) GetByID(ctx context.Context, id uint64) (*model.ConfigHistory, error) {
	var h model.ConfigHistory
	query := `SELECT id, config_id, namespace, ` + "`group`" + `, data_id, content, content_md5, version, op_type, op_by, created_at
			  FROM config_histories WHERE id=?`
	err := s.db.GetContext(ctx, &h, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *HistoryStore) GetLatestByConfigID(ctx context.Context, configID uint64) (*model.ConfigHistory, error) {
	var h model.ConfigHistory
	query := `SELECT id, config_id, namespace, ` + "`group`" + `, data_id, content, content_md5, version, op_type, op_by, created_at
			  FROM config_histories WHERE config_id=? ORDER BY version DESC LIMIT 1`
	err := s.db.GetContext(ctx, &h, query, configID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}