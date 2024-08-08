# GubGub

Yet another in-memory Go PubSub library.
I started to develop what is now GubGub in one of my personal projects but I soon found myself using it in other completely unrelated projects and I thought it could be a nice thing to share.
It might be very tailored to my personal preferences but my focus was concurrency safety and ease of usage.

## Getting started

```sh
go get -u gitlab.com/naterciom/gubgub
```

## Example

```Go
package main

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/naterciom/gubgub"
)

type MyMessage struct {
	Name string
}

func consumer(msg MyMessage) {
    fmt.Printf("Hello %s", msg.Name)
}

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	topic := gubgub.NewAsyncTopic[MyMessage](ctx)

	topic.Subscribe(gubgub.Forever(consumer))

    // The AsyncTopic doesn't wait for the subscriber to be registered so, for the purposes of this
    // example, we sleep on it.
	time.Sleep(time.Millisecond)

	topic.Publish(MyMessage{Name: "John Smith"})

	<-ctx.Done()
}
```

## Topic Benchmarks

So far GubGub implements 2 kinds of topics:

  * **SyncTopic** - Publishing blocks until the message was delivered. 
  Subscribers speed and number **will** have a direct impact the publishing performance.
  Under the right conditions (few and fast subscribers) this is the most performant topic.

  * **AsyncTopic** - Publishing schedules the message to be eventually delivered. 
  Subscribers speed and number **will not** directly impact the publishing perfomance at the cost of some publishing overhead.
  This is generally the most scalable topic.

The following benchmarks are just for reference on how the number of subscribers and their speed impact the publishing performance:

```
BenchmarkAsyncTopic_Publish/10_NoOp_Subscribers-8         	 2047338	       498.7 ns/op
BenchmarkAsyncTopic_Publish/100_NoOp_Subscribers-8        	 3317646	       535.0 ns/op
BenchmarkAsyncTopic_Publish/1K_NoOp_Subscribers-8         	 3239110	       578.9 ns/op
BenchmarkAsyncTopic_Publish/10K_NoOp_Subscribers-8        	 1871702	       691.2 ns/op
BenchmarkAsyncTopic_Publish/10_Slow_Subscribers-8         	 2615269	       433.4 ns/op
BenchmarkAsyncTopic_Publish/20_Slow_Subscribers-8         	 3127874	       470.4 ns/op
BenchmarkSyncTopic_Publish/10_NoOp_Subscribers-8          	24740354	        59.69 ns/op
BenchmarkSyncTopic_Publish/100_NoOp_Subscribers-8         	 4135681	       488.9 ns/op
BenchmarkSyncTopic_Publish/1K_NoOp_Subscribers-8          	  474122	      4320 ns/op
BenchmarkSyncTopic_Publish/10K_NoOp_Subscribers-8         	   45790	     35583 ns/op
BenchmarkSyncTopic_Publish/10_Slow_Subscribers-8          	  357253	      3393 ns/op
BenchmarkSyncTopic_Publish/20_Slow_Subscribers-8          	  179725	      6688 ns/op
```
