package gubgub

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncTopic_SinglePublisherSingleSubscriber(t *testing.T) {
	const msgCount = 10

	topic := NewSyncTopic[int]()

	var feedback []int

	err := topic.Subscribe(func(i int) bool {
		feedback = append(feedback, i)
		return true
	})
	require.NoError(t, err)

	for i := range msgCount {
		require.NoError(t, topic.Publish(i))
	}

	assert.Len(t, feedback, msgCount, "missing some feedback values")
}

func TestSyncTopic_MultiPublishersMultiSubscribers(t *testing.T) {
	const (
		subCount = 10
		pubCount = 10
		msgCount = pubCount * 100 // total messages to publish (delivered to EACH subscriber)
	)

	topic := NewSyncTopic[int]()

	var feedbackCounter int

	for range subCount {
		err := topic.Subscribe(func(i int) bool {
			feedbackCounter++
			return true
		})
		require.NoError(t, err)
	}

	toDeliver := make(chan int, msgCount)
	for i := range msgCount {
		toDeliver <- i
	}
	close(toDeliver)

	var wg sync.WaitGroup

	wg.Add(pubCount)
	for range pubCount {
		go func() {
			defer wg.Done()
			for msg := range toDeliver {
				require.NoError(t, topic.Publish(msg))
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, msgCount*subCount, feedbackCounter)
}
