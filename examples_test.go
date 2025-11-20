package gubgub_test

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nmoniz/gubgub"
)

func ExampleAsyncTopic() {
	latch := make(chan struct{})
	topic := gubgub.NewAsyncTopic[string](gubgub.WithOnSubscribe(func() { close(latch) }))
	defer topic.Close() // It's ok to close a topic multiple times

	receiver := make(chan string, 3) // closed later

	_ = topic.Subscribe(gubgub.Forever(func(msg string) {
		receiver <- strings.ToUpper(msg)
	}))

	<-latch // Wait for the subscribe to register

	_ = topic.Publish("aaa")
	_ = topic.Publish("bbb")
	_ = topic.Publish("ccc")

	topic.Close() // Close topic in order to wait for all messages to be delivered

	close(receiver) // Now it's safe to close the receiver channel since no more messages will be delivered

	msgLst := make([]string, 0, 3)
	for msg := range receiver {
		msgLst = append(msgLst, msg)
	}

	sort.Strings(msgLst) // Because publish order is not guaranteed with the AsyncTopic (yet) we need to sort results for consistent output

	for _, msg := range msgLst {
		fmt.Println(msg)
	}

	// Output: AAA
	// BBB
	// CCC
}

func ExampleSyncTopic() {
	topic := gubgub.NewSyncTopic[string]()
	defer topic.Close()

	receiver := make(chan string)
	defer close(receiver)

	_ = topic.Subscribe(gubgub.Buffered(gubgub.Forever(func(msg string) {
		receiver <- msg
	})))

	_ = topic.Publish("aaa")
	_ = topic.Publish("bbb")
	_ = topic.Publish("ccc")

	topic.Close()

	// Notice how despite the subscriber is blocked trying to push to an unbuffered channel it doesn't
	// block publishing to a SyncTopic thanks to the Buffered subscriber. Messages are considered
	// delivered even if they are in the buffer which is why Close returns.

	for range 3 {
		fmt.Println(<-receiver)
	}

	// Output: aaa
	// bbb
	// ccc
}
