package gubgub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithOnClose(t *testing.T) {
	type closable interface {
		Close()
	}

	feedback := make(chan struct{}, 1)
	defer close(feedback)
	onClose := WithOnClose(func() { feedback <- struct{}{} })

	testCases := []struct {
		name  string
		topic closable
	}{
		{
			name:  "sync topic",
			topic: NewSyncTopic[int](onClose),
		},
		{
			name:  "async topic",
			topic: NewAsyncTopic[int](onClose),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			go tc.topic.Close()

			select {
			case <-feedback:
				break

			case <-testTimer(t, time.Second).C:
				t.Fatalf("expected feedback by now")
			}
		})
	}
}

func TestWithOnSubscribe(t *testing.T) {
	const totalSub = 10

	feedback := make(chan int, totalSub)
	defer close(feedback)

	var counter int
	onSubscribe := WithOnSubscribe(func() {
		counter++
		feedback <- counter
	})

	testCases := []struct {
		name  string
		topic Subscribable[int]
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
			counter = 0

			for range totalSub {
				tc.topic.Subscribe(func(i int) bool { return true })
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
		})
	}
}
