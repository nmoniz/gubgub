package gubgub

// Forever wraps a subscriber that will never stop consuming messages.
// This helps avoiding subscribers that always return TRUE.
func Forever[T any](fn func(T)) Subscriber[T] {
	return func(msg T) bool {
		fn(msg)
		return true
	}
}

// Once wraps a subscriber that will consume only one message.
// This helps avoiding subscribers that always return FALSE.
func Once[T any](fn func(T)) Subscriber[T] {
	return func(t T) bool {
		fn(t)
		return false
	}
}

// NoOp creates a subscriber that does absolutely nothing forever.
// This is mostly useful for testing.
func NoOp[T any]() Subscriber[T] {
	return func(_ T) bool { return true }
}

// Buffered returns a subscriber that buffers messages if they can't be delivered immediately.
// There is no artificial limit to how many items can be buffered. This is bounded only by
// available memory.
// This is useful if message publishing is surge prone and message processing is slow or
// unpredictable (for example: subscriber makes network request).
// Message average processing rate must still be higher than the average message publishing rate
// otherwise it will eventually lead to memory issues. You will need to find a better strategy to
// deal with such scenario.
func Buffered[T any](subscriber Subscriber[T]) Subscriber[T] {
	unsubscribe := make(chan struct{}) // closed by the worker
	ready := make(chan struct{})       // closed by the worker
	messages := make(chan T)           // closed by the forwarder
	work := make(chan T)               // closed by the middleman

	// Worker calls the actual subscriber. It notifies the middleman that it's ready for the next
	// message via the ready channel and then reads from the work channel.
	go func() {
		for w := range work {
			if !subscriber(w) {
				close(unsubscribe)
				close(ready)
				return
			}
			ready <- struct{}{}
		}
	}()

	// Middleman that handles buffering. When the worker notifies that it is ready for the next
	// message it will check if there is buffered messages and push the next one immediately or
	// else push it when the next message arrives.
	go func() {
		defer close(work)

		idling := true // so that the first message can go straight to the consumer

		q := make([]T, 0, 1)

		for {
			select {
			case msg, more := <-messages:
				if !more {
					return
				}

				if idling {
					idling = false
					work <- msg
				} else {
					q = append(q, msg)
				}

			case _, more := <-ready:
				if !more {
					return
				}

				if len(q) > 0 {
					work <- q[0]
					q = q[1:]
				} else {
					idling = true
				}
			}

		}
	}()

	// forwarder just sends messages to the middleman or quits.
	return func(msg T) bool {
		select {
		case messages <- msg:
			return true
		case <-unsubscribe:
			close(messages)
			return false
		}
	}
}
