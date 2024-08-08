package gubgub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkAsyncTopic_Publish(b *testing.B) {
	for _, tc := range benchTestCase {
		b.Run(tc.Name, func(b *testing.B) {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			subscribersReady := make(chan struct{}, 1)
			defer close(subscribersReady)

			topicClosed := make(chan struct{}, 1)
			defer close(topicClosed)

			topic := NewAsyncTopic[int](ctx,
				WithOnSubscribe(func(count int) {
					if count == tc.Count {
						subscribersReady <- struct{}{}
					}
				}),
				WithOnClose(func() {
					topicClosed <- struct{}{}
				}),
			)

			for range tc.Count {
				require.NoError(b, topic.Subscribe(tc.Subscriber))
			}

			<-subscribersReady

			b.ResetTimer()

			for i := range b.N {
				_ = topic.Publish(i)
			}

			b.StopTimer()

			cancel()

			// This just helps leaving as few running Go routines as possible when the next round starts
			<-topicClosed
		})
	}

}
