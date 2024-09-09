package gubgub

import (
	"iter"
)

// Feed allows the usage of for/range loop to consume future published messages.
// The supporting subscriber will eventually be discarded after you exit the for loop.
func Feed[T any](t Subscribable[T], buffered bool) (iter.Seq[T], error) {
	feed := make(chan T)               // closed by the subscriber
	unsubscribe := make(chan struct{}) // closed by the iterator

	subscriber := func(msg T) bool {
		select {
		case feed <- msg:
			return true
		case <-unsubscribe:
			close(feed)
			return false
		}
	}

	if buffered {
		subscriber = Buffered(subscriber)
	}

	err := t.Subscribe(subscriber)
	if err != nil {
		return nil, err
	}

	// Iterator
	return func(yield func(T) bool) {
		defer close(unsubscribe)

		for msg := range feed {
			if !yield(msg) {
				return
			}
		}
	}, nil
}
