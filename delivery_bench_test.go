package gubgub

import "testing"

func BenchmarkSequentialDelivery(b *testing.B) {
	for _, tc := range deliveryCases {
		b.Run(tc.Name, func(b *testing.B) {
			subscribers := make([]Subscriber[int], 0, tc.Count)

			for range tc.Count {
				subscribers = append(subscribers, tc.Subscriber)
			}

			b.ResetTimer()

			for i := range b.N {
				b.StartTimer()
				subscribers = sequentialDelivery(i, subscribers)
				b.StopTimer()

				// replenish subscribers
				for len(subscribers) < tc.Count {
					subscribers = append(subscribers, tc.Subscriber)
				}
			}
		})
	}
}
