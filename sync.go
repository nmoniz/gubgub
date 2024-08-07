package gubgub

import "sync"

// SyncTopic allows any message T to be broadcast to subscribers. Publishing and Subscribing
// happens synchronously (block).
type SyncTopic[T any] struct {
	mu          sync.Mutex
	subscribers []Subscriber[T]
}

// NewSyncTopic creates a zero SyncTopic and return a pointer to it.
func NewSyncTopic[T any]() *SyncTopic[T] {
	return &SyncTopic[T]{}
}

// Publish broadcasts a message to all subscribers.
func (t *SyncTopic[T]) Publish(msg T) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	keepers := make([]Subscriber[T], 0, len(t.subscribers))

	for _, callback := range t.subscribers {
		keep := callback(msg)
		if keep {
			keepers = append(keepers, callback)
		}
	}

	t.subscribers = keepers

	return nil
}

// Subscribe adds a Subscriber func that will consume future published messages.
func (t *SyncTopic[T]) Subscribe(fn Subscriber[T]) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = append(t.subscribers, fn)

	return nil
}
