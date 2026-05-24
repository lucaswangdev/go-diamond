package watcher

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWatcherHub_Subscribe(t *testing.T) {
	hub := NewWatcherHub()

	ch1 := hub.Subscribe("ns1", "group1", "data1")
	ch2 := hub.Subscribe("ns1", "group1", "data1")

	assert.NotNil(t, ch1)
	assert.NotNil(t, ch2)
	assert.NotEqual(t, ch1, ch2)

	assert.Equal(t, 2, hub.WatcherCount())
}

func TestWatcherHub_Subscribe_DifferentKeys(t *testing.T) {
	hub := NewWatcherHub()

	ch1 := hub.Subscribe("ns1", "group1", "data1")
	ch2 := hub.Subscribe("ns1", "group1", "data2")

	assert.NotNil(t, ch1)
	assert.NotNil(t, ch2)
	assert.Equal(t, 2, hub.WatcherCount())
}

func TestWatcherHub_Unsubscribe(t *testing.T) {
	hub := NewWatcherHub()

	ch := hub.Subscribe("ns1", "group1", "data1")
	assert.Equal(t, 1, hub.WatcherCount())

	hub.Unsubscribe("ns1", "group1", "data1", ch)
	assert.Equal(t, 0, hub.WatcherCount())
}

func TestWatcherHub_Unsubscribe_Multiple(t *testing.T) {
	hub := NewWatcherHub()

	ch1 := hub.Subscribe("ns1", "group1", "data1")
	ch2 := hub.Subscribe("ns1", "group1", "data1")

	hub.Unsubscribe("ns1", "group1", "data1", ch1)
	assert.Equal(t, 1, hub.WatcherCount())

	hub.Unsubscribe("ns1", "group1", "data1", ch2)
	assert.Equal(t, 0, hub.WatcherCount())
}

func TestWatcherHub_Notify(t *testing.T) {
	hub := NewWatcherHub()

	ch := hub.Subscribe("ns1", "group1", "data1")

	done := make(chan struct{})
	go func() {
		select {
		case <-ch:
			close(done)
		case <-time.After(time.Second):
			t.Error("notify timed out")
		}
	}()

	hub.Notify("ns1", "group1", "data1")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("notify did not fire")
	}
}

func TestWatcherHub_Notify_MultipleSubscribers(t *testing.T) {
	hub := NewWatcherHub()

	ch1 := hub.Subscribe("ns1", "group1", "data1")
	ch2 := hub.Subscribe("ns1", "group1", "data1")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		<-ch1
		wg.Done()
	}()
	go func() {
		<-ch2
		wg.Done()
	}()

	hub.Notify("ns1", "group1", "data1")

	assert.True(t, waitTimeout(&wg, time.Second), "notify did not fire for all subscribers")
}

func TestWatcherHub_Notify_NoSubscriber(t *testing.T) {
	hub := NewWatcherHub()

	hub.Notify("ns1", "group1", "data1")

	assert.Equal(t, 0, hub.WatcherCount())
}

func TestWatcherHub_Notify_OtherNamespace(t *testing.T) {
	hub := NewWatcherHub()

	ch := hub.Subscribe("ns1", "group1", "data1")
	hub.Notify("ns2", "group1", "data1")

	select {
	case <-ch:
		t.Error("notify fired for wrong namespace")
	default:
	}
}

func TestWatcherHub_WatcherCount_Empty(t *testing.T) {
	hub := NewWatcherHub()
	assert.Equal(t, 0, hub.WatcherCount())
}

func TestWatcherHub_Concurrent(t *testing.T) {
	hub := NewWatcherHub()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			hub.Subscribe("ns", "group", "data")
			wg.Done()
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 10, hub.WatcherCount())
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}