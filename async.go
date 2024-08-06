package gubgub

import (
	"context"
	"fmt"
	"sync"
)

// AsyncTopic allows any message T to be broadcast to all subscribers.
type AsyncTopic[T any] struct {
	options AsyncTopicOptions

	mu     sync.RWMutex
	closed bool

	publishCh   chan T
	subscribeCh chan Subscriber[T]
}

// NewAsyncTopic creates a Topic that publishes messages asynchronously.
func NewAsyncTopic[T any](ctx context.Context, opts ...AsyncTopicOption) *AsyncTopic[T] {
	options := AsyncTopicOptions{
		onClose:     func() {},
		onSubscribe: func(count int) {},
	}

	for _, opt := range opts {
		opt(&options)
	}

	t := AsyncTopic[T]{
		options:     options,
		publishCh:   make(chan T, 1),
		subscribeCh: make(chan Subscriber[T]),
	}

	go t.run(ctx)

	return &t
}

func (t *AsyncTopic[T]) run(ctx context.Context) {
	defer close(t.publishCh)
	defer close(t.subscribeCh)
	defer t.close()

	var subscribers []Subscriber[T]

	for {
		select {
		case <-ctx.Done():
			return

		case newCallback := <-t.subscribeCh:
			subscribers = append(subscribers, newCallback)
			t.options.onSubscribe(len(subscribers))

		case msg := <-t.publishCh:
			keepers := make([]Subscriber[T], 0, len(subscribers))

			for _, callback := range subscribers {
				keep := callback(msg)
				if keep {
					keepers = append(keepers, callback)
				}
			}

			subscribers = keepers
		}
	}
}

// Publish broadcasts a msg to all subscribers.
func (t *AsyncTopic[T]) Publish(msg T) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return fmt.Errorf("async topic publish: %w", ErrTopicClosed)
	}

	t.publishCh <- msg
	return nil
}

// Subscribe adds a Subscriber func that will consume future published messages.
func (t *AsyncTopic[T]) Subscribe(fn Subscriber[T]) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return fmt.Errorf("async topic subscribe: %w", ErrTopicClosed)
	}

	t.subscribeCh <- fn
	return nil
}

func (t *AsyncTopic[T]) close() {
	t.mu.Lock()
	t.closed = true
	t.mu.Unlock()

	t.options.onClose()
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
