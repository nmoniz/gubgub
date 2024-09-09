package gubgub

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// SyncTopic is the simplest and most naive topic. It allows any message T to be broadcast to
// subscribers. Publishing and Subscribing happens synchronously (block).
type SyncTopic[T any] struct {
	options TopicOptions

	closed atomic.Bool

	mu          sync.Mutex
	subscribers []Subscriber[T]
}

// NewSyncTopic creates a zero SyncTopic and return a pointer to it.
func NewSyncTopic[T any](opts ...TopicOption) *SyncTopic[T] {
	options := TopicOptions{
		onClose:     func() {},
		onSubscribe: func(count int) {},
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &SyncTopic[T]{
		options: options,
	}
}

// Close will cause future Publish and Subscribe calls to return an error.
func (t *SyncTopic[T]) Close() {
	t.closed.Store(true)
	t.options.onClose()
}

// Publish broadcasts a message to all subscribers.
func (t *SyncTopic[T]) Publish(msg T) error {
	if t.closed.Load() {
		return fmt.Errorf("sync topic publish: %w", ErrTopicClosed)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = sequentialDelivery(msg, t.subscribers)

	return nil
}

// Subscribe adds a Subscriber func that will consume future published messages.
func (t *SyncTopic[T]) Subscribe(fn Subscriber[T]) error {
	if t.closed.Load() {
		return fmt.Errorf("sync topic subscribe: %w", ErrTopicClosed)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = append(t.subscribers, fn)
	t.options.onSubscribe(len(t.subscribers))

	return nil
}
