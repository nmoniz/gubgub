package gubgub

// sequentialDelivery delivers a message to each subscriber sequentially. For performance reasons
// this might mutate the subscribers slice inplace. Please overwrite it with the result of this
// call.
func sequentialDelivery[T any](msg T, subscribers []Subscriber[T]) []Subscriber[T] {
	last := len(subscribers) - 1
	next := 0

	for next <= last {
		if !subscribers[next](msg) {
			for last > next && !subscribers[last](msg) {
				last--
			}

			if last <= next {
				break
			}

			subscribers[next] = subscribers[last]
			last--
		}
		next++
	}

	return subscribers[:next]
}
