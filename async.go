package gubgub

import (
	"context"
	"fmt"
	"sync"
)

// AsyncTopic allows any message T to be broadcast to subscribers. Publishing as well as
// subscribing happens asynchronously. Due to the nature of async processes this cannot guarantee
// message delivery nor delivery order.
type AsyncTopic[T any] struct {
	options AsyncTopicOptions

	mu     sync.RWMutex
	closed bool

	publishCh   chan T
	subscribeCh chan Subscriber[T]
}

// NewAsyncTopic creates an AsyncTopic that will be closed when the given context is cancelled.
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

func (t *AsyncTopic[T]) close() {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		for range t.publishCh {
			// drain publishCh to release all read locks
		}
	}()

	go func() {
		defer wg.Done()
		for range t.subscribeCh {
			// drain subscribeCh to release all read locks
		}
	}()

	t.mu.Lock()
	t.closed = true // no more subscribing or publishing
	t.mu.Unlock()

	close(t.publishCh)
	close(t.subscribeCh)

	wg.Wait()

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
