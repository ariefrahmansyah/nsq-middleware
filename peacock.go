package peacock

import (
	"github.com/nsqio/go-nsq"
)

// Handler is an interface that objects can implement to be registered to serve as m
// in the Peacock m stack.
// HandleMessage should yield to the next m in the chain by invoking the next nsq.HandlerFunc
// passed in.
//
// If the Handler finishes the message, the next nsq.HandlerFunc is still be invoked.
type Handler interface {
	HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error
}

// WrapHandler converts a nsq.Handler into a peacock.Handler so it can be used as a Peacock
// m. The next nsq.HandlerFunc is automatically called after the Handler
// is executed.
func WrapHandler(handler nsq.Handler) Handler {
	return HandlerFunc(func(message *nsq.Message, next nsq.HandlerFunc) error {
		if err := handler.HandleMessage(message); err != nil {
			return err
		}
		return next(message)
	})
}

// HandlerFunc is an adapter to allow the use of ordinary functions as Peacock handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(message *nsq.Message, next nsq.HandlerFunc) error

func (handlerFunc HandlerFunc) HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error {
	return handlerFunc(message, next)
}

type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) HandleMessage(message *nsq.Message) error {
	return m.handler.HandleMessage(message, m.next.HandleMessage)
}

func buildMiddleware(handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		return emptyMiddleware()
	} else if len(handlers) > 1 {
		next = buildMiddleware(handlers[1:])
	} else {
		next = emptyMiddleware()
	}

	return middleware{handlers[0], &next}
}

func emptyMiddleware() middleware {
	return middleware{
		HandlerFunc(func(message *nsq.Message, next nsq.HandlerFunc) error { return nil }),
		&middleware{},
	}
}

// Peacock is a stack of Middleware Handlers that can be invoked as an nsq.Handler.
// Peacock middleware is evaluated in the order that they are added to the stack using
// the Use, UseHandler and UseHandlerFunc methods.
type Peacock struct {
	handlers   []Handler
	middleware middleware
}

// New returns a new Peacock instance with no middleware preconfigured.
func New(handlers ...Handler) *Peacock {
	return &Peacock{
		handlers:   handlers,
		middleware: buildMiddleware(handlers),
	}
}

func (peacock Peacock) HandleMessage(message *nsq.Message) error {
	return peacock.middleware.HandleMessage(message)
}

// Use adds a Handler onto the middleware stack. Handlers are invoked in the order they are added to a Peacock.
func (peacock *Peacock) Use(handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	peacock.handlers = append(peacock.handlers, handler)
	peacock.middleware = buildMiddleware(peacock.handlers)
}

// UseFunc adds a Peacock-style handler function onto the middleware stack.
func (peacock *Peacock) UseFunc(handlerFunc func(message *nsq.Message, next nsq.HandlerFunc) error) {
	peacock.Use(HandlerFunc(handlerFunc))
}

// UseHandler adds a nsq.Handler onto the middleware stack. Handlers are invoked in the order they are added to a Peacock.
func (peacock *Peacock) UseHandler(handler nsq.Handler) {
	peacock.Use(WrapHandler(handler))
}

// UseHandlerFunc adds a nsq.HandlerFunc-style handler function onto the middleware stack.
func (peacock *Peacock) UseHandlerFunc(handlerFunc func(message *nsq.Message) error) {
	peacock.UseHandler(nsq.HandlerFunc(handlerFunc))
}
