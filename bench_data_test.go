package gubgub

import "sync/atomic"

type benchSubscriberSetup struct {
	Name       string
	Count      int
	Subscriber Subscriber[int]
}

var publishCases = []benchSubscriberSetup{
	{
		Name:       "10 NoOp Subscribers",
		Count:      10,
		Subscriber: NoOp[int](),
	},
	{
		Name:       "100 NoOp Subscribers",
		Count:      100,
		Subscriber: NoOp[int](),
	},
	{
		Name:       "1K NoOp Subscribers",
		Count:      1000,
		Subscriber: NoOp[int](),
	},
	{
		Name:       "10K NoOp Subscribers",
		Count:      10000,
		Subscriber: NoOp[int](),
	},
	{
		Name:       "10 Slow Subscribers",
		Count:      10,
		Subscriber: Slow,
	},
	{
		Name:       "20 Slow Subscribers",
		Count:      20,
		Subscriber: Slow,
	},
}

func Slow(int) bool {
	for i := 0; i < 1000; i++ {
		// Just count to 1000
	}
	return true
}

var deliveryCases = []benchSubscriberSetup{
	{
		Name:       "10K Subscribers 0 unsubscribe",
		Count:      10000,
		Subscriber: NoOp[int](),
	},
	{
		Name:       "100K Subscribers 0 unsubscribe",
		Count:      100000,
		Subscriber: NoOp[int](),
	},
	{
		Name:       "10K Subscribers 1% unsubscribe",
		Count:      10000,
		Subscriber: Quiter(100),
	},
	{
		Name:       "100K Subscribers 1% unsubscribe",
		Count:      100000,
		Subscriber: Quiter(100),
	},
	{
		Name:       "10K Subscribers 2% unsubscribe",
		Count:      10000,
		Subscriber: Quiter(50),
	},
	{
		Name:       "100K Subscribers 2% unsubscribe",
		Count:      100000,
		Subscriber: Quiter(50),
	},
}

// Quiter returns a subscriber that unsubscribes nth calls.
func Quiter(nth int64) func(_ int) bool {
	var c atomic.Int64
	return func(_ int) bool {
		return c.Add(1)%nth != 0
	}
}
