package gubgub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequentialDelivery(t *testing.T) {
	const testMsg = 9786

	feedback := make([]int, 0, 3)

	subscribers := []Subscriber[int]{
		func(msg int) bool {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 1)
			return false
		},
		func(msg int) bool {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 2)
			return true
		},
		func(msg int) bool {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 3)
			return true
		},
	}

	nextSubscribers := sequentialDelivery(testMsg, subscribers)

	assert.Len(t, nextSubscribers, len(subscribers)-1, "expected to have one less subscriber")
	assert.Len(t, feedback, 3, "one or more subscriber was not called")

	finalSubscribers := sequentialDelivery(testMsg, nextSubscribers)

	assert.Len(t, finalSubscribers, len(nextSubscribers), "expected to have the same subscribers")
	assert.Len(t, feedback, 5, "one or more subscriber was not called")

	assertContainsExactlyN(t, 1, 1, feedback)
	assertContainsExactlyN(t, 2, 2, feedback)
	assertContainsExactlyN(t, 3, 2, feedback)
}

func assertContainsExactlyN[T comparable](t testing.TB, exp T, n int, slice []T) {
	t.Helper()

	var found int
	for _, v := range slice {
		if exp == v {
			found++
		}
	}

	if n > found {
		t.Errorf("contains too few '%v': expected %d but found %d", exp, n, found)
	} else if n < found {
		t.Errorf("contains too many '%v': expected %d but found %d", exp, n, found)
	}
}
