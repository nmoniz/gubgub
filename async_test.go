package gubgub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncTopic_SinglePublisherSingleSubscriber(t *testing.T) {
	const msgCount = 5

	topic := NewAsyncTopic[int](context.Background())

	feedback := make(chan int, msgCount)
	defer close(feedback)

	err := topic.Subscribe(func(i int) bool {
		feedback <- i
		return true
	})
	require.NoError(t, err)

	for i := 0; i < msgCount; i++ {
		require.NoError(t, topic.Publish(i))
	}

	count := 0
	for count < msgCount {
		select {
		case f := <-feedback:
			assert.Equalf(t, count, f, "expected to get %d but got %d instead", count, f)
			count++

		case <-time.After(time.Second):
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

	topic := NewAsyncTopic[int](context.Background())

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

	for range pubCount {
		go func() {
			for msg := range toDeliver {
				require.NoError(t, topic.Publish(msg))
			}
		}()
	}

	count := 0
	for count != expFeedbackCount {
		select {
		case <-feedback:
			count++

		case <-time.After(time.Second):
			t.Fatalf("expected %d feedback items by now but only got %d", expFeedbackCount, count)
		}
	}
}

func TestAsyncTopic_WithOnClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	feedback := make(chan struct{})
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

func TestAsyncTopic_SubscribeClosedTopicError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	feedback := make(chan struct{})
	defer close(feedback)

	topic := NewAsyncTopic[int](ctx, WithOnClose(func() { feedback <- struct{}{} }))

	require.NoError(t, topic.Subscribe(func(i int) bool { return true }))

	cancel() // this should close the topic, no more subscribers should be accepted

	select {
	case <-feedback:
		break

	case <-time.After(time.Second):
		t.Fatalf("expected feedback by now")
	}

	require.Error(t, ErrTopicClosed, topic.Subscribe(func(i int) bool { return true }))
}

func TestAsyncTopic_PublishClosedTopicError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	feedback := make(chan struct{})
	defer close(feedback)

	topic := NewAsyncTopic[int](ctx, WithOnClose(func() { feedback <- struct{}{} }))

	require.NoError(t, topic.Subscribe(func(i int) bool { return true }))

	require.NoError(t, topic.Publish(123))

	cancel() // this should close the topic, no more messages can be published

	select {
	case <-feedback:
		break

	case <-time.After(time.Second):
		t.Fatalf("expected feedback by now")
	}

	require.Error(t, ErrTopicClosed, topic.Publish(1))
}
