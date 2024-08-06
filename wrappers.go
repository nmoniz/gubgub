package gubgub

// Forever wrapper makes it more explicit that a subscriber will never stop consuming messages.
// This helps avoiding subscribers that always return true which, depending on their size, might
// not be immediately clear.
func Forever[T any](fn func(T)) Subscriber[T] {
	return func(msg T) bool {
		fn(msg)
		return true
	}
}
