package notifier

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/example/go-diamond/internal/store"
	"github.com/example/go-diamond/internal/watcher"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return sqlx.NewDb(db, "mysql"), mock
}

func TestDBNotifier_New(t *testing.T) {
	db, _ := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 2 * time.Second

	notifier := NewDBNotifier(configStore, hub, interval)

	assert.NotNil(t, notifier)
	assert.Equal(t, interval, notifier.interval)
}

func TestDBNotifier_Stop(t *testing.T) {
	db, _ := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 2 * time.Second

	notifier := NewDBNotifier(configStore, hub, interval)
	notifier.Stop()
}

func TestDBNotifier_check_NoUpdates(t *testing.T) {
	db, mock := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 2 * time.Second

	notifier := NewDBNotifier(configStore, hub, interval)
	notifier.lastCheck = time.Now().Add(-time.Hour)

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE updated_at > \\? AND is_deleted=0").
		WillReturnRows(sqlmock.NewRows([]string{}))

	notifier.check()

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBNotifier_check_WithUpdates(t *testing.T) {
	db, mock := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 2 * time.Second

	notifier := NewDBNotifier(configStore, hub, interval)
	notifier.lastCheck = time.Now().Add(-time.Hour)

	now := time.Now()
	rows := mock.NewRows([]string{"id", "namespace", "group", "data_id", "content", "content_md5", "format", "description", "version", "is_deleted", "created_by", "updated_by", "created_at", "updated_at"}).
		AddRow(1, "default", "DEFAULT_GROUP", "app.json", `{}`, "aaa", "json", "", 1, false, "admin", "admin", now, now)

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE updated_at > \\? AND is_deleted=0").
		WillReturnRows(rows)

	ch := hub.Subscribe("default", "DEFAULT_GROUP", "app.json")

	notifier.check()

	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Error("expected notification")
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBNotifier_check_IgnoresOldVersion(t *testing.T) {
	db, mock := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 2 * time.Second

	notifier := NewDBNotifier(configStore, hub, interval)
	notifier.lastCheck = time.Now().Add(-time.Hour)

	now := time.Now()
	rows := mock.NewRows([]string{"id", "namespace", "group", "data_id", "content", "content_md5", "format", "description", "version", "is_deleted", "created_by", "updated_by", "created_at", "updated_at"}).
		AddRow(1, "default", "DEFAULT_GROUP", "app.json", `{}`, "aaa", "json", "", 1, false, "admin", "admin", now, now)

	mock.ExpectQuery("SELECT id, namespace, `group`, data_id, content, content_md5, format, description, version, is_deleted, created_by, updated_by, created_at, updated_at FROM configs WHERE updated_at > \\? AND is_deleted=0").
		WillReturnRows(rows)

	key := watchKey{Namespace: "default", Group: "DEFAULT_GROUP", DataID: "app.json"}
	notifier.cache.Store(key, uint64(2))

	ch := hub.Subscribe("default", "DEFAULT_GROUP", "app.json")

	notifier.check()

	select {
	case <-ch:
		t.Error("did not expect notification for old version")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestDBNotifier_Start_ContextCancel(t *testing.T) {
	db, _ := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 50 * time.Millisecond

	notifier := NewDBNotifier(configStore, hub, interval)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	notifier.Start(ctx)
}

func TestDBNotifier_Start_Stop(t *testing.T) {
	db, _ := newMockDB(t)
	configStore := store.NewConfigStore(db)
	hub := watcher.NewWatcherHub()
	interval := 50 * time.Millisecond

	notifier := NewDBNotifier(configStore, hub, interval)

	ctx := context.Background()

	go func() {
		time.Sleep(20 * time.Millisecond)
		notifier.Stop()
	}()

	notifier.Start(ctx)
}