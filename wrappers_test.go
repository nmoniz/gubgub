package gubgub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuffered_Once(t *testing.T) {
	feedback := make(chan int, 1)
	s := Buffered(Once(func(i int) {
		feedback <- i // buffered channel means no blocking
	}))

	assert.True(t, s(1234))

	timeout := testTimer(t, time.Second)

	select {
	case i := <-feedback:
		assert.Equal(t, 1234, i)

	case <-timeout.C:
		t.Fatalf("expected feedback value by now")
	}

	assert.False(t, s(4321))
}

func TestBuffered_Forever(t *testing.T) {
	const msgCount = 100

	feedback := make(chan int)
	s := Buffered(Forever(func(i int) {
		feedback <- i // unbuffered channel creates choke point (blocks) to force buffering
	}))

	for i := range msgCount {
		assert.True(t, s(i))
	}

	timeout := testTimer(t, time.Second)

	var count int
	for count < msgCount {
		select {
		case <-feedback:
			count++

		case <-timeout.C:
			t.Fatalf("expected %d feedback values by now", msgCount)
		}
	}
}
