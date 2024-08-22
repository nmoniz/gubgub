package gubgub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequentialDelivery(t *testing.T) {
	const testMsg = 9786

	feedback := make([]int, 0, 3)

	subscribers := []Subscriber[int]{
		Once(func(msg int) {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 1)
		}),
		Once(func(msg int) {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 2)
		}),
		Once(func(msg int) {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 3)
		}),
		Forever(func(msg int) {
			assert.Equalf(t, testMsg, msg, "expected %d but got %d", testMsg, msg)
			feedback = append(feedback, 4)
		}),
	}

	nextSubscribers := sequentialDelivery(testMsg, subscribers)

	assert.Len(t, nextSubscribers, 1, "expected to have 1 subscriber")
	assert.Len(t, feedback, 4, "one or more subscriber was not called")

	finalSubscribers := sequentialDelivery(testMsg, nextSubscribers)

	assert.Len(t, finalSubscribers, len(nextSubscribers), "expected to have the same subscribers")
	assert.Len(t, feedback, 5, "one or more subscriber was not called")

	assertContainsExactlyN(t, 1, 1, feedback)
	assertContainsExactlyN(t, 2, 1, feedback)
	assertContainsExactlyN(t, 3, 1, feedback)
	assertContainsExactlyN(t, 4, 2, feedback)
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
