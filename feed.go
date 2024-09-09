package gubgub

import "iter"

// Feed allows the usage of for/range to consume future published messages.
// The supporting subscriber will eventually be discarded after you exit the for loop.
func Feed[T any](t Subscribable[T], buffered bool) iter.Seq[T] {
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

	t.Subscribe(subscriber)

	// Iterator
	return func(yield func(T) bool) {
		defer close(unsubscribe)

		for msg := range feed {
			if !yield(msg) {
				return
			}
		}
	}
}
