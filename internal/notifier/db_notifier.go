package notifier

import (
	"context"
	"sync"
	"time"

	"github.com/example/go-diamond/internal/store"
	"github.com/example/go-diamond/internal/watcher"
)

type DBNotifier struct {
	store     *store.ConfigStore
	hub       *watcher.WatcherHub
	cache     sync.Map
	interval  time.Duration
	stopCh    chan struct{}
	lastCheck time.Time
}

func NewDBNotifier(store *store.ConfigStore, hub *watcher.WatcherHub, interval time.Duration) *DBNotifier {
	return &DBNotifier{
		store:    store,
		hub:      hub,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (n *DBNotifier) Start(ctx context.Context) {
	n.lastCheck = time.Now()

	ticker := time.NewTicker(n.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.check()
		case <-n.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (n *DBNotifier) check() {
	configs, err := n.store.GetUpdatedConfigs(context.Background(), n.lastCheck)
	if err != nil {
		return
	}

	if len(configs) > 0 {
		n.lastCheck = time.Now()
	}

	for _, cfg := range configs {
		key := watchKey{
			Namespace: cfg.Namespace,
			Group:     cfg.Group,
			DataID:    cfg.DataID,
		}

		if v, ok := n.cache.Load(key); ok {
			if v.(uint64) >= cfg.Version {
				continue
			}
		}

		n.cache.Store(key, cfg.Version)
		n.hub.Notify(cfg.Namespace, cfg.Group, cfg.DataID)
	}
}

type watchKey struct {
	Namespace string
	Group     string
	DataID    string
}

func (n *DBNotifier) Stop() {
	close(n.stopCh)
}