# nsqg [![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/ariefrahmansyah/nsqg) [![CircleCI](https://circleci.com/gh/ariefrahmansyah/nsqg/tree/master.png?style=shield)](https://circleci.com/gh/ariefrahmansyah/nsqg/tree/master) [![Coverage Status](https://coveralls.io/repos/github/ariefrahmansyah/nsqg/badge.svg?branch=master)](https://coveralls.io/github/ariefrahmansyah/nsqg?branch=master) [![GoReportCard](https://goreportcard.com/badge/github.com/ariefrahmansyah/nsqg)](https://goreportcard.com/report/github.com/ariefrahmansyah/nsqg)

NSQG is middleware for your NSQ consumer handler function. It's heavily inspired by [negroni](https://github.com/urfave/negroni).

## Usage
We can create new middleware object that implement nsq.Handler interface and use it in NSQG object using Use method.

We also can use nsq.HandlerFunc that we already created before.

```go
// Create nsqg object.
nsqg := nsqg.New()

// Create middleware object that implement nsq.Handler interface.
type Middleware1 struct{}
func (m1 Middleware1) HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error {
	log.Printf("Middleware 1:\t%s\n", message.Body)
	return next(message)
}

// Use Middleware1.
nsqg.Use(Middleware1{})

// We may already have nsq.HandlerFunc.
func handlerFunc1(message *nsq.Message) error {
	log.Printf("Handler Func 1:\t%s\n\n", message.Body)
	message.Finish()
	return nil
}

// Let's use it too.
nsqg.UseHandlerFunc(handlerFunc1)

consumer, _ := nsq.NewConsumer(topicName, channelName, nsq.NewConfig())
consumer.AddHandler(nsqg)
consumer.ConnectToNSQD(nsqdAddress)
```
