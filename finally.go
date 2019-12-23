package finally

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
)

type FinallyHandler func()
type FinallyHandlerSig func(sig os.Signal)

func Wrap(handler FinallyHandler) func() {
	c := newHandlerContext()
	c.handler = handler
	return c.executeNoSig
}

func WrapSig(handler FinallyHandlerSig) func() {
	c := newHandlerContext()
	c.acceptArgument = true
	c.sigHandler = handler
	return c.executeNoSig
}

func RegisterShutdownHook(signals ...os.Signal) {
	ch := make(chan os.Signal)

	if signals == nil {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	signal.Notify(ch, signals...)

	go func() {
		sig := <-ch

		handlers.Range(func(key, value interface{}) bool {
			hc := value.(*handlerContext)
			hc.execute(&sig)
			return true
		})

		os.Exit(1)
	}()
}

func SetRecordStackTrace(v bool) {
	recordStackTrace = v
}

var seq int64
var handlers sync.Map
var recordStackTrace = true

type handlerContext struct {
	seq            int64
	isExecuted     int32
	handler        FinallyHandler
	sigHandler     FinallyHandlerSig
	acceptArgument bool
	stackTrace     []byte
}

func newHandlerContext() *handlerContext {
	c := new(handlerContext)
	newSeq := atomic.AddInt64(&seq, 1)
	c.seq = newSeq

	if recordStackTrace {
		c.stackTrace = debug.Stack()
	}

	handlers.Store(newSeq, c)
	return c
}

func (c *handlerContext) execute(sig *os.Signal) {
	isExecuted := atomic.AddInt32(&c.isExecuted, 1)
	if isExecuted > 1 {
		return
	}

	defer func() {
		handlers.Delete(c.seq)

		if err := recover(); err != nil {
			fmt.Println("Panic caught in finally handler:", err)
			if c.stackTrace != nil {
				fmt.Println(string(c.stackTrace))
			}
		}
	}()

	if c.acceptArgument {
		c.sigHandler(*sig)
	} else {
		c.handler()
	}
}

func (c *handlerContext) executeNoSig() {
	c.execute(nil)
}
