package watcher

import (
	"sync"
)

type watchKey struct {
	Namespace string
	Group     string
	DataID    string
}

type WatcherHub struct {
	mu       sync.RWMutex
	watchers map[watchKey][]chan struct{}
}

func NewWatcherHub() *WatcherHub {
	return &WatcherHub{
		watchers: make(map[watchKey][]chan struct{}),
	}
}

func (h *WatcherHub) Subscribe(namespace, group, dataID string) chan struct{} {
	key := watchKey{Namespace: namespace, Group: group, DataID: dataID}
	ch := make(chan struct{}, 1)

	h.mu.Lock()
	defer h.mu.Unlock()

	h.watchers[key] = append(h.watchers[key], ch)
	return ch
}

func (h *WatcherHub) Unsubscribe(namespace, group, dataID string, ch chan struct{}) {
	key := watchKey{Namespace: namespace, Group: group, DataID: dataID}

	h.mu.Lock()
	defer h.mu.Unlock()

	chans := h.watchers[key]
	for i, c := range chans {
		if c == ch {
			h.watchers[key] = append(chans[:i], chans[i+1:]...)
			break
		}
	}
}

func (h *WatcherHub) Notify(namespace, group, dataID string) {
	key := watchKey{Namespace: namespace, Group: group, DataID: dataID}

	h.mu.RLock()
	chans := h.watchers[key]
	h.mu.RUnlock()

	for _, ch := range chans {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (h *WatcherHub) WatcherCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, chans := range h.watchers {
		total += len(chans)
	}
	return total
}