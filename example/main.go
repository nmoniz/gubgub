package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nmoniz/gubgub"
)

func main() {
	topic := gubgub.NewAsyncTopic[string]()
	defer topic.Close()

	err := topic.Subscribe(gubgub.Forever(UpperCaser))
	if err != nil {
		log.Fatal(err)
	}

	err = topic.Subscribe(Countdown(3))
	if err != nil {
		log.Fatal(err)
	}

	err = topic.Subscribe(gubgub.Buffered(gubgub.Forever(Slow)))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		feed, err := gubgub.Feed(topic, false)
		if err != nil {
			log.Fatal(err)
		}

		for s := range feed {
			log.Printf("ForRange: %s", s)
		}
	}()

	fmt.Printf("Use 'Ctrl+C' to exit! Type a message followed by 'Enter' to publish it:\n")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		topic.Publish(scanner.Text())
	}
}

func UpperCaser(input string) {
	log.Printf("UpperCaser: %s", strings.ToUpper(input))
}

func Countdown(count int) gubgub.Subscriber[string] {
	return func(_ string) bool {
		count--
		if count > 0 {
			log.Printf("Countdown: %d...", count)
			return true
		} else {
			log.Print("Countdown: I'm out!")
			return false
		}
	}
}

func Slow(input string) {
	time.Sleep(time.Second * 3)
	log.Printf("Slow: %s", input)
}
