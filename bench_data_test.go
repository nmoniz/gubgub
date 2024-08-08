package gubgub

type benchSubscriberSetup struct {
	Name       string
	Count      int
	Subscriber Subscriber[int]
}

var benchTestCase = []benchSubscriberSetup{
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
