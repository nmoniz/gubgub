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
