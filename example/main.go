package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gitlab.com/naterciom/gubgub"
)

func main() {
	topic := gubgub.NewAsyncTopic[string]()
	defer topic.Close()

	topic.Subscribe(gubgub.Forever(UpperCaser))

	topic.Subscribe(Countdown(3))

	topic.Subscribe(gubgub.Buffered(gubgub.Forever(Slow)))

	fmt.Printf("Use 'Ctrl+C' to exit! Type messages followed by 'Enter' to publish them:\n")
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
