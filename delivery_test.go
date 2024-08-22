package gubgub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequentialDelivery(t *testing.T) {
	const testMsg = 9786

	feedback := make([]int, 0, 3)

	subscribers := []Subscriber[int]{
		func(i int) bool {
			assert.Equalf(t, testMsg, i, "expected %d but got %d", testMsg, i)
			feedback = append(feedback, 1)
			return true
		},
		func(i int) bool {
			assert.Equalf(t, testMsg, i, "expected %d but got %d", testMsg, i)
			feedback = append(feedback, 2)
			return false
		},
		func(i int) bool {
			assert.Equalf(t, testMsg, i, "expected %d but got %d", testMsg, i)
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
}
