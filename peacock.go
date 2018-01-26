package peacock

import (
	"github.com/nsqio/go-nsq"
)

type Handler interface {
	HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error
}

func WrapHandler(handler nsq.Handler) Handler {
	return HandlerFunc(func(message *nsq.Message, next nsq.HandlerFunc) error {
		if err := handler.HandleMessage(message); err != nil {
			return err
		}
		return next(message)
	})
}

type HandlerFunc func(message *nsq.Message, next nsq.HandlerFunc) error

func (handlerFunc HandlerFunc) HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error {
	return handlerFunc(message, next)
}

type Middleware struct {
	handler Handler
	next    *Middleware
}

func (middleware Middleware) HandleMessage(message *nsq.Message) error {
	return middleware.handler.HandleMessage(message, middleware.next.HandleMessage)
}

func buildMiddleware(handlers []Handler) Middleware {
	var next Middleware

	if len(handlers) == 0 {
		return emptyMiddleware()
	} else if len(handlers) > 1 {
		next = buildMiddleware(handlers[1:])
	} else {
		next = emptyMiddleware()
	}

	return Middleware{handlers[0], &next}
}

func emptyMiddleware() Middleware {
	return Middleware{
		HandlerFunc(func(message *nsq.Message, next nsq.HandlerFunc) error { return nil }),
		&Middleware{},
	}
}

type Peacock struct {
	handlers   []Handler
	middleware Middleware
}

func New(handlers ...Handler) *Peacock {
	return &Peacock{
		handlers:   handlers,
		middleware: buildMiddleware(handlers),
	}
}

func (peacock Peacock) HandleMessage(message *nsq.Message) error {
	return peacock.middleware.HandleMessage(message)
}

func (peacock *Peacock) Use(handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	peacock.handlers = append(peacock.handlers, handler)
	peacock.middleware = buildMiddleware(peacock.handlers)
}

func (peacock *Peacock) UseFunc(handlerFunc func(message *nsq.Message, next nsq.HandlerFunc) error) {
	peacock.Use(HandlerFunc(handlerFunc))
}

func (peacock *Peacock) UseHandler(handler nsq.Handler) {
	peacock.Use(WrapHandler(handler))
}

func (peacock *Peacock) UseHandlerFunc(handlerFunc func(message *nsq.Message) error) {
	peacock.UseHandler(nsq.HandlerFunc(handlerFunc))
}
