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

// NewSyncTopic creates a SyncTopic with the specified options.
func NewSyncTopic[T any](opts ...TopicOption) *SyncTopic[T] {
	t := &SyncTopic[T]{}

	t.SetOptions(opts...)

	return t
}

// Close will prevent further publishing and subscribing.
func (t *SyncTopic[T]) Close() {
	t.closed.Store(true)
	t.options.TriggerClose()
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
	t.options.TriggerSubscribe()

	return nil
}

func (t *SyncTopic[T]) SetOptions(opts ...TopicOption) {
	t.options.Apply(opts...)
}
