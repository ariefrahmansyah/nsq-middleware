package nsqmiddleware

import "github.com/nsqio/go-nsq"

// Handler is an interface that objects can implement to be registered to serve as middleware
// in the NSQM middleware stack.
// HandleMessage should yield to the next middleware in the chain by invoking the next nsq.HandlerFunc
// passed in.
//
// If the Handler finishes the message, the next nsq.HandlerFunc is still be invoked.
type Handler interface {
	HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error
}

// WrapHandler converts a nsq.Handler into a nsqm.Handler so it can be used as a NSQM middleware.
// The next nsq.HandlerFunc is automatically called after the Handler is executed.
func WrapHandler(handler nsq.Handler) Handler {
	return HandlerFunc(func(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
		if err := handler.HandleMessage(message); err != nil {
			return err
		}
		return next(message)
	})
}

// HandlerFunc is an adapter to allow the use of ordinary functions as NSQM handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error

func (handlerFunc HandlerFunc) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	return handlerFunc(topic, channel, message, next)
}

type middleware struct {
	topic   string
	channel string
	handler Handler
	next    *middleware
}

func (m middleware) HandleMessage(message *nsq.Message) error {
	return m.handler.HandleMessage(m.topic, m.channel, message, m.next.HandleMessage)
}

func buildMiddleware(topic, channel string, handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		return emptyMiddleware()
	} else if len(handlers) > 1 {
		next = buildMiddleware(topic, channel, handlers[1:])
	} else {
		next = emptyMiddleware()
	}

	return middleware{topic, channel, handlers[0], &next}
}

func emptyMiddleware() middleware {
	return middleware{
		topic:   "",
		channel: "",
		handler: HandlerFunc(func(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error { return nil }),
		next:    &middleware{},
	}
}

// NSQM is a stack of Middleware Handlers that can be invoked as an nsq.Handler.
// NSQM middleware is evaluated in the order that they are added to the stack using
// the Use, UseHandler and UseHandlerFunc methods.
type NSQM struct {
	topic      string
	channel    string
	handlers   []Handler
	middleware middleware
}

// New returns a new NSQM instance with no middleware preconfigured.
func New(topic, channel string, handlers ...Handler) *NSQM {
	return &NSQM{
		topic:      topic,
		channel:    channel,
		handlers:   handlers,
		middleware: buildMiddleware(topic, channel, handlers),
	}
}

func (nsqm NSQM) HandleMessage(message *nsq.Message) error {
	return nsqm.middleware.HandleMessage(message)
}

// Use adds a Handler onto the middleware stack. Handlers are invoked in the order they are added to a NSQM.
func (nsqm *NSQM) Use(handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	nsqm.handlers = append(nsqm.handlers, handler)
	nsqm.middleware = buildMiddleware(nsqm.topic, nsqm.channel, nsqm.handlers)
}

// UseFunc adds a NSQM-style handler function onto the middleware stack.
func (nsqm *NSQM) UseFunc(handlerFunc func(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error) {
	nsqm.Use(HandlerFunc(handlerFunc))
}

// UseHandler adds a nsq.Handler onto the middleware stack. Handlers are invoked in the order they are added to a NSQM.
func (nsqm *NSQM) UseHandler(handler nsq.Handler) {
	nsqm.Use(WrapHandler(handler))
}

// UseHandlerFunc adds a nsq.HandlerFunc-style handler function onto the middleware stack.
func (nsqm *NSQM) UseHandlerFunc(handlerFunc func(message *nsq.Message) error) {
	nsqm.UseHandler(nsq.HandlerFunc(handlerFunc))
}
