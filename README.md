# GubGub

Another Go PubSub library.

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

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	topic := gubgub.NewAsyncTopic[MyMessage](ctx)

	topic.Subscribe(gubgub.Forever(func(msg MyMessage) {
		fmt.Printf("Hello %s", msg.Name)
	}))

	topic.Publish(MyMessage{Name: "John Smith"})

	<-ctx.Done()
}
```
