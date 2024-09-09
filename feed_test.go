package gubgub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFeed_Topics(t *testing.T) {
	const msgCount = 10

	subscriberReady := make(chan struct{}, 1)
	defer close(subscriberReady)

	onSubscribe := WithOnSubscribe(func(count int) {
		subscriberReady <- struct{}{}
	})

	testCases := []struct {
		name  string
		topic Topic[int]
	}{
		{
			name:  "sync topic",
			topic: NewSyncTopic[int](onSubscribe),
		},
		{
			name:  "async topic",
			topic: NewAsyncTopic[int](onSubscribe),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			feedback := make(chan int)
			go func() {
				for i := range Feed(tc.topic, false) {
					feedback <- i
				}
			}()

			go func() {
				<-subscriberReady
				for i := range msgCount {
					require.NoError(t, tc.topic.Publish(i))
				}
			}()

			var counter int
			timeout := testTimer(t, time.Second)
			for counter < msgCount {
				select {
				case <-feedback:
					counter++
				case <-timeout.C:
					t.Fatalf("expected %d feedback values by now but only got %d", msgCount, counter)
				}
			}
		})
	}

}
