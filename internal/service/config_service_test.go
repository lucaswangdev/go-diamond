package service

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/example/go-diamond/internal/store"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return sqlx.NewDb(db, "mysql"), mock
}

func TestNewConfigService(t *testing.T) {
	db, _ := newMockDB(t)
	configStore := store.NewConfigStore(db)
	historyStore := store.NewHistoryStore(db)

	svc := NewConfigService(configStore, historyStore)
	assert.NotNil(t, svc)
}

func TestMd5Hash(t *testing.T) {
	hash := md5Hash("test content")
	assert.Len(t, hash, 32)
	assert.Equal(t, "9473fdd0d880a43c21b7778d34872157", md5Hash("test content"))
}

func TestMd5Hash_DifferentContent(t *testing.T) {
	hash1 := md5Hash("content1")
	hash2 := md5Hash("content2")
	assert.NotEqual(t, hash1, hash2)
}

func TestMd5Hash_EmptyString(t *testing.T) {
	hash := md5Hash("")
	assert.Len(t, hash, 32)
}