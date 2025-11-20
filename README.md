# GubGub

[![Go Report Card](https://goreportcard.com/badge/github.com/nmoniz/gubgub)](https://goreportcard.com/report/github.com/nmoniz/gubgub)
[![Coverage Status](https://coveralls.io/repos/github/nmoniz/gubgub/badge.svg?branch=main)](https://coveralls.io/github/nmoniz/gubgub?branch=main)

Yet another in-memory Go PubSub library.
I started to develop what is now GubGub in one of my personal projects but I soon found myself using it in other completely unrelated projects and I thought it could be a nice thing to share.

## Getting started

```sh
go get github.com/nmoniz/gubgub@latest
```

## Example

We'll be ignoring errors for code brevity!

```Go
package main

import (
 "context"
 "fmt"
 "time"

 "github.com/nmoniz/gubgub"
)

type MyMessage struct {
 Name string
}

func consumer(msg MyMessage) {
 fmt.Printf("Hello %s", msg.Name)
}

func main() {
 topic := gubgub.NewAsyncTopic[MyMessage]()
 defer topic.Close() // Returns after all messages are delivered

 _ = topic.Subscribe(gubgub.Forever(consumer))

 // The AsyncTopic doesn't wait for the subscriber to be registered so, for the purposes of this
 // example, we sleep on it.
 time.Sleep(time.Millisecond)

 _ = topic.Publish(MyMessage{Name: "John Smith"}) // Returns immediately
}
```

## Topics

Topics are what this is all about.
You publish to a topic and you subscribe to a topic.
That is it.

A `Subscriber` is just a callback `func` with a signature: `func[T any](message T) bool`.
A `Subscriber` can unsubscribe by returning false.
A `message` is considered delivered when all subscribers have been called and returned for that message.

If you `Publish` a message successfully (did not get an error) then you can be sure the message will be delivered before any call to `Close` returns.

Topics are meant to live as long as the application but you should call the `Close` method upon shutdown to fulfill the publishing promise.
Use the `WithOnClose` option when creating the topic to perform any extra clean up you might need to do if the topic is closed.

GubGub offers 2 kinds of topics:

* **SyncTopic** - Publishing blocks until the message was delivered to all subscribers.
  Subscribing blocks until the subscriber is registered.

* **AsyncTopic** - Publishing schedules the message to be eventually delivered.
  Subscribing schedules a subscriber to be eventually registered.
  Only message delivery is guaranteed.

The type of topic does not relate to how messages are actually delivered.
Currently we deliver messages sequentially (each subscriber gets the message one after the other).

## Benchmarks

* **SyncTopic** - Subscribers speed and number **will** have a direct impact the publishing performance.
  Under the right conditions (few and fast subscribers) this is the most performant topic.

* **AsyncTopic** - Subscribers speed and number **will not** directly impact the publishing performance at the cost of some publishing overhead.
  This is generally the most scalable topic.

The following benchmarks are just for topic comparison regarding how the number of subscribers and their speed can impact the publishing performance:

```
BenchmarkAsyncTopic_Publish/10_NoOp_Subscribers-8           2047338        498.7 ns/op
BenchmarkAsyncTopic_Publish/100_NoOp_Subscribers-8          3317646        535.0 ns/op
BenchmarkAsyncTopic_Publish/1K_NoOp_Subscribers-8           3239110        578.9 ns/op
BenchmarkAsyncTopic_Publish/10K_NoOp_Subscribers-8          1871702        691.2 ns/op
BenchmarkAsyncTopic_Publish/10_Slow_Subscribers-8           2615269        433.4 ns/op
BenchmarkAsyncTopic_Publish/20_Slow_Subscribers-8           3127874        470.4 ns/op
BenchmarkSyncTopic_Publish/10_NoOp_Subscribers-8           24740354         59.69 ns/op
BenchmarkSyncTopic_Publish/100_NoOp_Subscribers-8           4135681        488.9 ns/op
BenchmarkSyncTopic_Publish/1K_NoOp_Subscribers-8             474122       4320 ns/op
BenchmarkSyncTopic_Publish/10K_NoOp_Subscribers-8             45790      35583 ns/op
BenchmarkSyncTopic_Publish/10_Slow_Subscribers-8             357253       3393 ns/op
BenchmarkSyncTopic_Publish/20_Slow_Subscribers-8             179725       6688 ns/op
```
