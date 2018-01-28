package nsqmiddleware

import (
	"log"
	"os"
	"runtime"

	"github.com/nsqio/go-nsq"
)

const panicText = "PANIC: %s\n%s"

// Recovery is a NSQ-Middleware that recovers from any panics.
type Recovery struct {
	Logger     ILogger
	PrintStack bool
	StackAll   bool
	StackSize  int
}

// NewRecovery returns a new instance of Recovery.
func NewRecovery() *Recovery {
	return &Recovery{
		Logger:     log.New(os.Stdout, "[nsqm] ", 0),
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

func (recovery *Recovery) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, recovery.StackSize)
			stack = stack[:runtime.Stack(stack, recovery.StackAll)]

			recovery.Logger.Printf(panicText, err, stack)
		}
	}()

	return next(message)
}
