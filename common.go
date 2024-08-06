package gubgub

// Subscriber is a func that processes a message and returns true if it should continue processing more messages.
type Subscriber[T any] func(T) bool

// Topic is just a convenience interface you can expect all topics to implement.
type Topic[T any] interface {
	Publishable[T]
	Subscribable[T]
}

type Publishable[T any] interface {
	Publish(msg T) error
}

type Subscribable[T any] interface {
	Subscribe(Subscriber[T]) error
}
