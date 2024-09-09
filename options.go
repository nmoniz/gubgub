package gubgub

// TopicOptions holds common options for topics.
type TopicOptions struct {
	// onClose is called after the Topic is closed and all messages have been delivered.
	onClose func()

	// onSubscribe is called after a new subscriber is regitered.
	onSubscribe func(count int)
}

type TopicOption func(*TopicOptions)

func WithOnClose(fn func()) TopicOption {
	return func(opts *TopicOptions) {
		opts.onClose = fn
	}
}

func WithOnSubscribe(fn func(count int)) TopicOption {
	return func(opts *TopicOptions) {
		opts.onSubscribe = fn
	}
}
