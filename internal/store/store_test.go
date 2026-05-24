package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/example/go-diamond/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return sqlx.NewDb(db, "mysql"), mock
}

func TestConfigStore_Get(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	rows := mock.NewRows([]string{"id", "namespace", "group", "data_id", "content", "content_md5", "format", "description", "version", "is_deleted", "created_by", "updated_by", "created_at", "updated_at"}).
		AddRow(1, "default", "DEFAULT_GROUP", "app.json", `{"key":"value"}`, "abc123", "json", "", 1, false, "admin", "admin", time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE namespace=\\? AND `group`=\\? AND data_id=\\? AND is_deleted=0").
		WithArgs("default", "DEFAULT_GROUP", "app.json").
		WillReturnRows(rows)

	cfg, err := store.Get(context.Background(), "default", "DEFAULT_GROUP", "app.json")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, uint64(1), cfg.ID)
	assert.Equal(t, "default", cfg.Namespace)
	assert.Equal(t, "DEFAULT_GROUP", cfg.Group)
	assert.Equal(t, "app.json", cfg.DataID)
	assert.Equal(t, `{"key":"value"}`, cfg.Content)
}

func TestConfigStore_Get_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE namespace=\\? AND `group`=\\? AND data_id=\\? AND is_deleted=0").
		WithArgs("default", "DEFAULT_GROUP", "notexist.json").
		WillReturnError(sql.ErrNoRows)

	cfg, err := store.Get(context.Background(), "default", "DEFAULT_GROUP", "notexist.json")
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestConfigStore_Create(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	mock.ExpectExec("INSERT INTO configs").
		WithArgs("default", "DEFAULT_GROUP", "app.json", `{"key":"value"}`, "abc123", "json", "", uint64(1), "admin", "admin").
		WillReturnResult(sqlmock.NewResult(1, 1))

	cfg := &model.Config{
		Namespace:   "default",
		Group:       "DEFAULT_GROUP",
		DataID:      "app.json",
		Content:     `{"key":"value"}`,
		ContentMD5:  "abc123",
		Format:      "json",
		Description: "",
		Version:     1,
		CreatedBy:   "admin",
		UpdatedBy:   "admin",
	}

	err := store.Create(context.Background(), cfg)
	require.NoError(t, err)
}

func TestConfigStore_Update(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	mock.ExpectExec("UPDATE configs SET content=\\?, content_md5=\\?, format=\\?, description=\\?, version=\\?, updated_by=\\?, updated_at=NOW\\(\\) WHERE namespace=\\? AND `group`=\\? AND data_id=\\? AND is_deleted=0").
		WithArgs(`{"key":"updated"}`, "def456", "json", "", uint64(2), "admin", "default", "DEFAULT_GROUP", "app.json").
		WillReturnResult(sqlmock.NewResult(0, 1))

	cfg := &model.Config{
		ID:          1,
		Namespace:   "default",
		Group:       "DEFAULT_GROUP",
		DataID:      "app.json",
		Content:     `{"key":"updated"}`,
		ContentMD5:  "def456",
		Format:      "json",
		Description: "",
		Version:     2,
		UpdatedBy:   "admin",
	}

	err := store.Update(context.Background(), cfg)
	require.NoError(t, err)
}

func TestConfigStore_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	mock.ExpectExec("UPDATE configs SET is_deleted=1, updated_by=\\?, updated_at=NOW\\(\\) WHERE namespace=\\? AND `group`=\\? AND data_id=\\? AND is_deleted=0").
		WithArgs("admin", "default", "DEFAULT_GROUP", "app.json").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.Delete(context.Background(), "default", "DEFAULT_GROUP", "app.json", "admin")
	require.NoError(t, err)
}

func TestConfigStore_List(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM configs WHERE namespace=\\? AND `group`=\\? AND is_deleted=0").
		WithArgs("default", "DEFAULT_GROUP").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := mock.NewRows([]string{"id", "namespace", "group", "data_id", "content", "content_md5", "format", "description", "version", "is_deleted", "created_by", "updated_by", "created_at", "updated_at"}).
		AddRow(1, "default", "DEFAULT_GROUP", "app1.json", `{}`, "aaa", "json", "", 1, false, "admin", "admin", time.Now(), time.Now()).
		AddRow(2, "default", "DEFAULT_GROUP", "app2.json", `{}`, "bbb", "json", "", 1, false, "admin", "admin", time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE namespace=\\? AND `group`=\\? AND is_deleted=0 ORDER BY updated_at DESC LIMIT \\? OFFSET \\?").
		WithArgs("default", "DEFAULT_GROUP", 20, 0).
		WillReturnRows(rows)

	configs, total, err := store.List(context.Background(), "default", "DEFAULT_GROUP", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, configs, 2)
}

func TestConfigStore_GetUpdatedConfigs(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewConfigStore(db)

	rows := mock.NewRows([]string{"id", "namespace", "group", "data_id", "content", "content_md5", "format", "description", "version", "is_deleted", "created_by", "updated_by", "created_at", "updated_at"}).
		AddRow(1, "default", "DEFAULT_GROUP", "app.json", `{}`, "aaa", "json", "", 1, false, "admin", "admin", time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE updated_at > \\? AND is_deleted=0").
		WillReturnRows(rows)

	since := time.Now().Add(-time.Hour)
	configs, err := store.GetUpdatedConfigs(context.Background(), since)
	require.NoError(t, err)
	assert.Len(t, configs, 1)
}

func TestHistoryStore_Create(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewHistoryStore(db)

	mock.ExpectExec("INSERT INTO config_histories").
		WithArgs(uint64(1), "default", "DEFAULT_GROUP", "app.json", `{"key":"value"}`, "abc123", uint64(1), "create", "admin").
		WillReturnResult(sqlmock.NewResult(1, 1))

	history := &model.ConfigHistory{
		ConfigID:   1,
		Namespace:  "default",
		Group:      "DEFAULT_GROUP",
		DataID:     "app.json",
		Content:    `{"key":"value"}`,
		ContentMD5: "abc123",
		Version:    1,
		OpType:     "create",
		OpBy:       "admin",
	}

	err := store.Create(context.Background(), history)
	require.NoError(t, err)
}

func TestHistoryStore_List(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewHistoryStore(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM config_histories WHERE config_id=\\?").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := mock.NewRows([]string{"id", "config_id", "namespace", "group", "data_id", "content", "content_md5", "version", "op_type", "op_by", "created_at"}).
		AddRow(1, 1, "default", "DEFAULT_GROUP", "app.json", `{}`, "aaa", 1, "create", "admin", time.Now()).
		AddRow(2, 1, "default", "DEFAULT_GROUP", "app.json", `{}`, "bbb", 2, "update", "admin", time.Now())

	mock.ExpectQuery("SELECT id, config_id, namespace, `group`, data_id, content, content_md5, version, op_type, op_by, created_at FROM config_histories WHERE config_id=\\? ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
		WithArgs(uint64(1), 20, 0).
		WillReturnRows(rows)

	histories, total, err := store.List(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, histories, 2)
}

func TestHistoryStore_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewHistoryStore(db)

	rows := mock.NewRows([]string{"id", "config_id", "namespace", "group", "data_id", "content", "content_md5", "version", "op_type", "op_by", "created_at"}).
		AddRow(1, 1, "default", "DEFAULT_GROUP", "app.json", `{}`, "aaa", 1, "create", "admin", time.Now())

	mock.ExpectQuery("SELECT id, config_id, namespace, `group`, data_id, content, content_md5, version, op_type, op_by, created_at FROM config_histories WHERE id=\\?").
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	history, err := store.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, history)
	assert.Equal(t, uint64(1), history.ID)
}

func TestHistoryStore_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewHistoryStore(db)

	mock.ExpectQuery("SELECT id, config_id, namespace, `group`, data_id, content, content_md5, version, op_type, op_by, created_at FROM config_histories WHERE id=\\?").
		WithArgs(uint64(999)).
		WillReturnError(sql.ErrNoRows)

	history, err := store.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, history)
}

func TestHistoryStore_GetLatestByConfigID(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewHistoryStore(db)

	rows := mock.NewRows([]string{"id", "config_id", "namespace", "group", "data_id", "content", "content_md5", "version", "op_type", "op_by", "created_at"}).
		AddRow(2, 1, "default", "DEFAULT_GROUP", "app.json", `{}`, "bbb", 2, "update", "admin", time.Now())

	mock.ExpectQuery("SELECT id, config_id, namespace, `group`, data_id, content, content_md5, version, op_type, op_by, created_at FROM config_histories WHERE config_id=\\? ORDER BY version DESC LIMIT 1").
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	history, err := store.GetLatestByConfigID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, history)
	assert.Equal(t, uint64(2), history.Version)
}

func TestHistoryStore_GetLatestByConfigID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	store := NewHistoryStore(db)

	mock.ExpectQuery("SELECT id, config_id, namespace, `group`, data_id, content, content_md5, version, op_type, op_by, created_at FROM config_histories WHERE config_id=\\? ORDER BY version DESC LIMIT 1").
		WithArgs(uint64(999)).
		WillReturnError(sql.ErrNoRows)

	history, err := store.GetLatestByConfigID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, history)
}