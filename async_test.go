package gubgub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncTopic_SinglePublisherSingleSubscriber(t *testing.T) {
	const msgCount = 10

	subscriberReady := make(chan struct{}, 1)
	defer close(subscriberReady)

	topic := NewAsyncTopic[int](context.Background(), WithOnSubscribe(func(count int) {
		subscriberReady <- struct{}{}
	}))

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
	timeout := time.After(time.Second)

	for count < msgCount {
		select {
		case <-feedback:
			count++

		case <-timeout:
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

	subscribersReady := make(chan struct{}, 1)
	defer close(subscribersReady)

	topic := NewAsyncTopic[int](context.Background(), WithOnSubscribe(func(count int) {
		if count == subCount {
			subscribersReady <- struct{}{}
		}
	}))

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
	timeout := time.After(time.Second)

	for count != expFeedbackCount {
		select {
		case <-feedback:
			count++

		case <-timeout:
			t.Fatalf("expected %d feedback items by now but only got %d", expFeedbackCount, count)
		}
	}
}

func TestAsyncTopic_WithOnClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	feedback := make(chan struct{}, 1)
	defer close(feedback)

	_ = NewAsyncTopic[int](ctx, WithOnClose(func() { feedback <- struct{}{} }))

	cancel()

	select {
	case <-feedback:
		break

	case <-time.After(time.Second):
		t.Fatalf("expected feedback by now")
	}
}

func TestAsyncTopic_WithOnSubscribe(t *testing.T) {
	const totalSub = 10

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	feedback := make(chan int, totalSub)
	defer close(feedback)

	topic := NewAsyncTopic[int](ctx, WithOnSubscribe(func(count int) { feedback <- count }))

	for range totalSub {
		topic.Subscribe(func(i int) bool { return true })
	}

	count := 0
	timeout := time.After(time.Second)

	for count < totalSub {
		select {
		case c := <-feedback:
			count++
			assert.Equal(t, count, c, "expected %d but got %d instead", count, c)

		case <-timeout:
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			feedback := make(chan struct{}, 1)
			defer close(feedback)

			topic := NewAsyncTopic[int](ctx, WithOnClose(func() { feedback <- struct{}{} }))

			cancel() // this should close the topic, no more messages can be published

			select {
			case <-feedback:
				break

			case <-time.After(time.Second):
				t.Fatalf("expected feedback by now")
			}

			tc.assertFn(topic)
		})
	}
}

func TestAsyncTopic_AllPublishedBeforeClosedAreDeliveredAfterClosed(t *testing.T) {
	const msgCount = 10

	subscriberReady := make(chan struct{}, 1)
	defer close(subscriberReady)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := NewAsyncTopic[int](ctx, WithOnSubscribe(func(count int) {
		subscriberReady <- struct{}{}
	}))

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

	cancel()

	values := make(map[int]struct{}, msgCount)
	timeout := time.After(time.Second)

	for len(values) < msgCount {
		select {
		case f := <-feedback:
			values[f] = struct{}{}

		case <-timeout:
			t.Fatalf("expected %d unique feedback values by now but only got %d", msgCount, len(values))
		}
	}
}
