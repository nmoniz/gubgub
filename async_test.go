package gubgub

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncTopic_SinglePublisherSingleSubscriber(t *testing.T) {
	const msgCount = 10

	onSubscribe, subscriberReady := withNotifyOnNthSubscriber(t, 1)
	topic := newTestAsyncTopic[int](t, onSubscribe)

	feedback := make(chan struct{}, msgCount)
	defer close(feedback)

	err := topic.Subscribe(func(i int) bool {
		feedback <- struct{}{}
		return true
	})
	require.NoError(t, err)

	<-subscriberReady

	for i := range msgCount {
		require.NoError(t, topic.Publish(i))
	}

	count := 0
	timeout := testTimer(t, time.Second)

	for count < msgCount {
		select {
		case <-feedback:
			count++

		case <-timeout.C:
			t.Fatalf("expected %d feedback items by now but only got %d", msgCount, count)
		}
	}
}

func TestAsyncTopic_MultiPublishersMultiSubscribers(t *testing.T) {
	const (
		subCount = 10
		pubCount = 10
		msgCount = pubCount * 100 // total messages to publish (delivered to EACH subscriber)
	)

	onSubscribe, subscribersReady := withNotifyOnNthSubscriber(t, subCount)
	topic := newTestAsyncTopic[int](t, onSubscribe)

	expFeedbackCount := msgCount * subCount
	feedback := make(chan int, expFeedbackCount)
	defer close(feedback)

	for range subCount {
		err := topic.Subscribe(func(i int) bool {
			feedback <- i
			return true
		})
		require.NoError(t, err)
	}

	toDeliver := make(chan int, msgCount)
	for i := range msgCount {
		toDeliver <- i
	}
	close(toDeliver)

	<-subscribersReady

	for range pubCount {
		go func() {
			for msg := range toDeliver {
				require.NoError(t, topic.Publish(msg))
			}
		}()
	}

	count := 0
	timeout := testTimer(t, time.Second)

	for count != expFeedbackCount {
		select {
		case <-feedback:
			count++

		case <-timeout.C:
			t.Fatalf("expected %d feedback items by now but only got %d", expFeedbackCount, count)
		}
	}
}

func TestAsyncTopic_WithOnClose(t *testing.T) {
	feedback := make(chan struct{}, 1)
	defer close(feedback)

	topic := NewAsyncTopic[int](WithOnClose(func() { feedback <- struct{}{} }))

	topic.Close()

	select {
	case <-feedback:
		break

	case <-testTimer(t, time.Second).C:
		t.Fatalf("expected feedback by now")
	}
}

func TestAsyncTopic_WithOnSubscribe(t *testing.T) {
	const totalSub = 10

	feedback := make(chan int, totalSub)
	defer close(feedback)

	topic := NewAsyncTopic[int](WithOnSubscribe(func(count int) { feedback <- count }))

	for range totalSub {
		topic.Subscribe(func(i int) bool { return true })
	}

	count := 0
	timeout := testTimer(t, time.Second)

	for count < totalSub {
		select {
		case c := <-feedback:
			count++
			assert.Equal(t, count, c, "expected %d but got %d instead", count, c)

		case <-timeout.C:
			t.Fatalf("expected %d feedback items by now but only got %d", totalSub, count)
		}
	}
}

func TestAsyncTopic_ClosedTopicError(t *testing.T) {
	testCases := []struct {
		name     string
		assertFn func(*AsyncTopic[int])
	}{
		{
			name: "publishing returns error",
			assertFn: func(topic *AsyncTopic[int]) {
				assert.Error(t, ErrTopicClosed, topic.Publish(1))
			},
		},
		{
			name: "subscribing returns error",
			assertFn: func(topic *AsyncTopic[int]) {
				assert.Error(t, ErrTopicClosed, topic.Subscribe(func(i int) bool { return true }))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			feedback := make(chan struct{}, 1)
			defer close(feedback)

			topic := NewAsyncTopic[int]()

			topic.Close() // this should close the topic, no more messages can be published

			tc.assertFn(topic)
		})
	}
}

func TestAsyncTopic_AllPublishedBeforeClosedAreDeliveredAfterClosed(t *testing.T) {
	const msgCount = 10

	onSubscribe, subscriberReady := withNotifyOnNthSubscriber(t, 1)
	topic := newTestAsyncTopic[int](t, onSubscribe)

	feedback := make(chan int) // unbuffered will cause choke point for publishers
	defer close(feedback)

	err := topic.Subscribe(func(i int) bool {
		feedback <- i
		return true
	})
	require.NoError(t, err)

	<-subscriberReady

	for i := range msgCount {
		require.NoError(t, topic.Publish(i))
	}

	go topic.Close()

	values := make(map[int]struct{}, msgCount)
	timeout := testTimer(t, time.Second)

	for len(values) < msgCount {
		select {
		case f := <-feedback:
			values[f] = struct{}{}

		case <-timeout.C:
			t.Fatalf("expected %d unique feedback values by now but only got %d", msgCount, len(values))
		}
	}
}

func testTimer(t testing.TB, d time.Duration) *time.Timer {
	t.Helper()

	timer := time.NewTimer(d)
	t.Cleanup(func() {
		timer.Stop()
	})

	return timer
}

func newTestAsyncTopic[T any](t testing.TB, opts ...TopicOption) *AsyncTopic[T] {
	t.Helper()
	topic := NewAsyncTopic[T](opts...)
	t.Cleanup(topic.Close)
	return topic
}

func withNotifyOnNthSubscriber(t testing.TB, n int64) (TopicOption, <-chan struct{}) {
	t.Helper()

	notify := make(chan struct{}, 1)
	t.Cleanup(func() {
		close(notify)
	})

	var counter atomic.Int64

	return WithOnSubscribe(func(count int) {
		c := counter.Add(1)
		if c == n {
			notify <- struct{}{}
		}
	}), notify
}
