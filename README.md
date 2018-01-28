# nsq-middleware [![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/ariefrahmansyah/nsq-middleware) [![CircleCI](https://circleci.com/gh/ariefrahmansyah/nsq-middleware/tree/master.png?style=shield)](https://circleci.com/gh/ariefrahmansyah/nsq-middleware/tree/master) [![Coverage Status](https://coveralls.io/repos/github/ariefrahmansyah/nsq-middleware/badge.svg?branch=master)](https://coveralls.io/github/ariefrahmansyah/nsq-middleware?branch=master) [![GoReportCard](https://goreportcard.com/badge/github.com/ariefrahmansyah/nsq-middleware)](https://goreportcard.com/report/github.com/ariefrahmansyah/nsq-middleware)

NSQ-Middleware is middleware for your NSQ consumer handler function. It's heavily inspired by [negroni](https://github.com/urfave/negroni).

## Bundled Middleware
1. Recovery
2. Logger
3. Prometheus

## Usage
We can create new middleware object that implement nsq.Handler interface and use it in NSQ-Middleware object using Use method.

We also can use nsq.HandlerFunc that we already created before.

```go
// Create NSQ-Middleware object.
nsqMid := nsqmiddleware.New()

// Optional: Use three bundled middleware.
nsqMid.Use(nsqm.NewRecovery())
nsqMid.Use(nsqm.NewLogger())
nsqMid.Use(nsqm.NewPrometheus())

// Create middleware object that implement nsq.Handler interface.
type Middleware1 struct{}
func (m1 Middleware1) HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error {
	log.Printf("Middleware 1:\t%s\n", message.Body)
	return next(message)
}

// Use Middleware1.
nsqMid.Use(Middleware1{})

// We may already have nsq.HandlerFunc.
func handlerFunc1(message *nsq.Message) error {
	log.Printf("Handler Func 1:\t%s\n\n", message.Body)
	message.Finish()
	return nil
}

// Let's use it too.
nsqMid.UseHandlerFunc(handlerFunc1)

consumer, _ := nsq.NewConsumer(topicName, channelName, nsq.NewConfig())
consumer.AddHandler(nsqMid)
consumer.ConnectToNSQD(nsqdAddress)
```

Check [example package](https://github.com/ariefrahmansyah/nsq-middleware/blob/master/example) for more usage examples.
