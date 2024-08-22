package gubgub

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkSyncTopic_Publish(b *testing.B) {
	for _, tc := range publishCases {
		b.Run(tc.Name, func(b *testing.B) {
			topic := NewSyncTopic[int]()

			for range tc.Count {
				require.NoError(b, topic.Subscribe(tc.Subscriber))
			}

			b.ResetTimer()

			for i := range b.N {
				_ = topic.Publish(i)
			}
		})
	}
}
