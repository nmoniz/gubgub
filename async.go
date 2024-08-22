package gubgub

import (
	"context"
	"fmt"
	"iter"
	"sync"
)

// AsyncTopic allows any message T to be broadcast to subscribers. Publishing as well as
// subscribing happens asynchronously.
// This guarantees that every published message will be delivered but does NOT guarantee delivery
// order.
// In the unlikely scenario where subscribers are being queued very aggressively it is possible
// that some might never actually receive any message. Subscriber registration order is also not
// guaranteed.
type AsyncTopic[T any] struct {
	options AsyncTopicOptions

	mu     sync.RWMutex
	closed bool

	publishCh   chan T
	subscribeCh chan Subscriber[T]
}

// NewAsyncTopic creates an AsyncTopic that will be closed when the given context is cancelled.
// After closed calls to Publish or Subscribe will return an error.
func NewAsyncTopic[T any](ctx context.Context, opts ...AsyncTopicOption) *AsyncTopic[T] {
	options := AsyncTopicOptions{
		onClose:     func() {},          // Called after the Topic is closed and all messages have been delivered.
		onSubscribe: func(count int) {}, // Called everytime a new subscriber is added
	}

	for _, opt := range opts {
		opt(&options)
	}

	t := AsyncTopic[T]{
		options:     options,
		publishCh:   make(chan T, 1),
		subscribeCh: make(chan Subscriber[T], 1),
	}

	go t.closer(ctx)
	go t.run()

	return &t
}

func (t *AsyncTopic[T]) closer(ctx context.Context) {
	<-ctx.Done()

	t.mu.Lock()
	t.closed = true // no more subscribing or publishing
	t.mu.Unlock()

	close(t.publishCh)
	close(t.subscribeCh)
}

func (t *AsyncTopic[T]) run() {
	defer t.options.onClose()

	var subscribers []Subscriber[T]

	var drainedSubscribe, drainedPublish bool
	for !drainedSubscribe || !drainedPublish {
		select {
		case newCallback, more := <-t.subscribeCh:
			if !more {
				drainedSubscribe = true
				break
			}

			subscribers = append(subscribers, newCallback)
			t.options.onSubscribe(len(subscribers))

		case msg, more := <-t.publishCh:
			if !more {
				drainedPublish = true
				break
			}

			subscribers = sequentialDelivery(msg, subscribers)
		}
	}
}

// Publish broadcasts a msg to all subscribers.
func (t *AsyncTopic[T]) Publish(msg T) error {
	t.mu.RLock()

	if t.closed {
		t.mu.RUnlock()
		return fmt.Errorf("async topic publish: %w", ErrTopicClosed)
	}

	go func() {
		t.publishCh <- msg
		t.mu.RUnlock()
	}()

	return nil
}

// Subscribe adds a Subscriber func that will consume future published messages.
func (t *AsyncTopic[T]) Subscribe(fn Subscriber[T]) error {
	t.mu.RLock()

	if t.closed {
		t.mu.RUnlock()
		return fmt.Errorf("async topic subscribe: %w", ErrTopicClosed)
	}

	go func() {
		t.subscribeCh <- fn
		t.mu.RUnlock()
	}()

	return nil
}

// Feed allows the usage of for/range to consume future published messages. The supporting subscriber will eventually be discarded after you exit the for loop.
func (t *AsyncTopic[T]) Feed() iter.Seq[T] {
	feed := make(chan T, 1)
	done := make(chan struct{})

	t.Subscribe(func(msg T) bool {
		select {
		case feed <- msg:
			return true
		case <-done:
			close(feed)
			return false
		}
	})

	return func(yield func(T) bool) {
		defer close(done)

		for msg := range feed {
			if !yield(msg) {
				return
			}
		}
	}
}

type AsyncTopicOptions struct {
	onClose     func()
	onSubscribe func(count int)
}

type AsyncTopicOption func(*AsyncTopicOptions)

func WithOnClose(fn func()) AsyncTopicOption {
	return func(opts *AsyncTopicOptions) {
		opts.onClose = fn
	}
}

func WithOnSubscribe(fn func(count int)) AsyncTopicOption {
	return func(opts *AsyncTopicOptions) {
		opts.onSubscribe = fn
	}
}
