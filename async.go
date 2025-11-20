package gubgub

import (
	"fmt"
	"sync"
)

// AsyncTopic allows any message T to be broadcast to subscribers. Publishing as well as
// subscribing happens asynchronously (as non-blocking as possible).
// Closing the topic guarantees that published message will be delivered and no further messages
// nor subscribers will be accepted.
// Delivery order is NOT guaranteed.
type AsyncTopic[T any] struct {
	options TopicOptions

	mu      sync.RWMutex
	closing bool
	closed  chan struct{}

	publishCh   chan T
	subscribeCh chan Subscriber[T]
}

// NewAsyncTopic creates an AsyncTopic.
func NewAsyncTopic[T any](opts ...TopicOption) *AsyncTopic[T] {
	options := TopicOptions{
		onClose:     func() {}, // Called after the Topic is closed and all messages have been delivered.
		onSubscribe: func() {}, // Called everytime a new subscriber is added
	}

	for _, opt := range opts {
		opt(&options)
	}

	t := AsyncTopic[T]{
		options:     options,
		closed:      make(chan struct{}),
		publishCh:   make(chan T, 1),
		subscribeCh: make(chan Subscriber[T], 1),
	}

	go t.run()

	return &t
}

// Close terminates background go routines and prevents further publishing and subscribing. All
// published messages are garanteed to be delivered once Close returns. This is idempotent and
// thread safe.
func (t *AsyncTopic[T]) Close() {
	t.mu.Lock()
	if t.closing {
		// Multiple go routines attempted to close this topic. All should wait for the topic to be
		// closed before returning.
		t.mu.Unlock()
		<-t.closed
		return
	}

	t.closing = true // no more subscribing or publishing
	t.mu.Unlock()

	close(t.publishCh)
	close(t.subscribeCh)

	<-t.closed
}

func (t *AsyncTopic[T]) run() {
	defer close(t.closed)
	defer t.options.onClose()

	var subscribers []Subscriber[T]

	defer func() {
		// There is only one way to get here: the topic is now closing!
		// Because both `subscribeCh` and `publishCh` channels are closed when the topic is closed
		// we can assume this will always eventually return.
		// This will deliver any potential queued message thus fulfilling the message delivery
		// promise.
		for msg := range t.publishCh {
			subscribers = sequentialDelivery(msg, subscribers)
		}
	}()

	for {
		select {
		case newCallback, more := <-t.subscribeCh:
			if !more {
				return
			}

			subscribers = append(subscribers, newCallback)
			t.options.onSubscribe()

		case msg, more := <-t.publishCh:
			if !more {
				return
			}

			subscribers = sequentialDelivery(msg, subscribers)
		}
	}
}

// Publish broadcasts a msg to all subscribers asynchronously.
func (t *AsyncTopic[T]) Publish(msg T) error {
	t.mu.RLock()

	// We hold the Read lock until we are done with publishing to avoid panic due to a closed channel.

	if t.closing {
		t.mu.RUnlock()
		return fmt.Errorf("async topic publish: %w", ErrTopicClosed)
	}

	go func() {
		t.publishCh <- msg
		t.mu.RUnlock()
	}()

	return nil
}

// Subscribe registers a Subscriber func asynchronously.
func (t *AsyncTopic[T]) Subscribe(fn Subscriber[T]) error {
	t.mu.RLock()

	if t.closing {
		t.mu.RUnlock()
		return fmt.Errorf("async topic subscribe: %w", ErrTopicClosed)
	}

	go func() {
		t.subscribeCh <- fn
		t.mu.RUnlock()
	}()

	return nil
}
