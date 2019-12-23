package main

import (
	"fmt"
	"time"

	"github.com/chrisww/finally"
)

func sub() {
	f := finally.Wrap(func() {
		fmt.Println("Triggered")
	})

	defer f()

	time.Sleep(1 * time.Second)
}

func main() {
	finally.RegisterShutdownHook()

	sub()

	time.Sleep(1 * time.Second)
}
