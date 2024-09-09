package gubgub

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkAsyncTopic_Publish(b *testing.B) {
	for _, tc := range publishCases {
		b.Run(tc.Name, func(b *testing.B) {
			onSubscribe, subscribersReady := withNotifyOnNthSubscriber(b, int64(tc.Count))
			topic := NewAsyncTopic[int](onSubscribe)
			b.Cleanup(topic.Close)

			for range tc.Count {
				require.NoError(b, topic.Subscribe(tc.Subscriber))
			}

			<-subscribersReady

			b.ResetTimer()

			for i := range b.N {
				_ = topic.Publish(i)
			}

			b.StopTimer()

			topic.Close()
		})
	}

}
