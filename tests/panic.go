package main

import (
	"fmt"
	"time"

	"github.com/chrisww/finally"
)

func main() {
	finally.RegisterShutdownHook()

	finally.Wrap(func() {
		panic("Test Panic")
	})

	finally.Wrap(func() {
		fmt.Println("Triggered")
	})

	finally.Wrap(func() {
		panic("Test Panic")
	})

	time.Sleep(3 * time.Second)
}
