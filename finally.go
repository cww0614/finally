// finally is a package helping to enhance cleanup handlers which is
// usually used with defer. By wrapping functions with finally.Wrap,
// cleanup handlers are guaranteed to be executed even when the
// program is terminated by SIGTERM.
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

// Wrap wraps a function of type FinallyHandler and returns a function
// to invoke the input function, which can be used in defer. The input
// function will be called when the program receives shutdown signals
// or with the returned function. The input function is guranteed to
// be invoked only once.
func Wrap(handler FinallyHandler) func() {
	c := newHandlerContext()
	c.handler = handler
	return c.executeNoSig
}

// WrapSig does exactly the same thing as Wrap, except that the
// function can accept an argument representing the signal that the
// program received. If the wrapping function is called, the signal
// will be nil.
func WrapSig(handler FinallyHandlerSig) func() {
	c := newHandlerContext()
	c.acceptArgument = true
	c.sigHandler = handler
	return c.executeNoSig
}

// RegisterShutdownHook registers signal handlers for signals
// specified in the arguments. If no arguments are specified,
// os.Interrupt and syscall.SIGTERM are used.
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

// SetRecordStackTrace set the config about whether the stacktrace
// should be saved. If it is set to true, when a panic happenes in a
// finally handler, the stacktrace will be shown.
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
