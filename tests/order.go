package main

import (
	"fmt"
	"time"

	"github.com/chrisww/finally"
)

func main() {
	finally.RegisterShutdownHook()

	f1 := finally.Wrap(func() {
		fmt.Println("Triggered1")
	})

	f2 := finally.Wrap(func() {
		fmt.Println("Triggered2")
	})

	f3 := finally.Wrap(func() {
		fmt.Println("Triggered3")
	})

	defer f1()
	defer f2()
	defer f3()

	time.Sleep(3 * time.Second)
}
