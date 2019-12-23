package main

import (
	"fmt"
	"time"

	"github.com/chrisww/finally"
)

func main() {
	finally.RegisterShutdownHook()

	f := finally.Wrap(func() {
		fmt.Println("Triggered")
	})

	defer f()

	time.Sleep(3 * time.Second)
}
