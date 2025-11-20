package gubgub

import "sync"

// TopicOptions holds common options for topics.
type TopicOptions struct {
	mu sync.Mutex

	// onClose is called after the Topic is closed and all messages have been delivered. Even
	// though you might call Close multiple times, topics are effectively closed only once thus
	// this should be called only once.
	onClose func()

	// onSubscribe is called after a new subscriber is regitered.
	onSubscribe func()
}

func (to *TopicOptions) TriggerClose() {
	to.mu.Lock()
	defer to.mu.Unlock()

	if to.onClose == nil {
		return
	}

	to.onClose()
}

func (to *TopicOptions) TriggerSubscribe() {
	to.mu.Lock()
	defer to.mu.Unlock()

	if to.onSubscribe == nil {
		return
	}

	to.onSubscribe()
}

func (to *TopicOptions) Apply(opts ...TopicOption) {
	to.mu.Lock()
	defer to.mu.Unlock()

	for _, opt := range opts {
		opt(to)
	}
}

type TopicOption func(*TopicOptions)

func WithOnClose(fn func()) TopicOption {
	return func(opts *TopicOptions) {
		if opts.onClose == nil {
			opts.onClose = fn
		} else {
			oldFn := opts.onClose // preserve previous onClose handler
			opts.onClose = func() {
				fn()
				oldFn()
			}
		}
	}
}

func WithOnSubscribe(fn func()) TopicOption {
	return func(opts *TopicOptions) {
		if opts.onSubscribe == nil {
			opts.onSubscribe = fn
		} else {
			oldFn := opts.onSubscribe // preserve previous onSubscribe handler
			opts.onSubscribe = func() {
				fn()
				oldFn()
			}
		}
	}
}
