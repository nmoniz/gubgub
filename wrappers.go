package gubgub

// Forever wrapper makes it more explicit that a subscriber will never stop consuming messages.
// This helps avoiding subscribers that always return TRUE.
func Forever[T any](fn func(T)) Subscriber[T] {
	return func(msg T) bool {
		fn(msg)
		return true
	}
}

// Once returns a subscriber that will consume only one message.
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
