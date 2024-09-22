package gubgub

import (
	"sync"
)

// TopicOptions holds common options for topics.
type TopicOptions struct {
	// onClose is called after the Topic is closed and all messages have been delivered. Even
	// though you might call Close multiple times, topics are effectively closed only once thus
	// this should be called only once.
	onClose func()

	// onSubscribe is called after a new subscriber is regitered.
	onSubscribe func()
}

type TopicOption func(*TopicOptions)

func WithOnClose(fn func()) TopicOption {
	return func(opts *TopicOptions) {
		opts.onClose = sync.OnceFunc(fn)
	}
}

func WithOnSubscribe(fn func()) TopicOption {
	return func(opts *TopicOptions) {
		opts.onSubscribe = fn
	}
}
