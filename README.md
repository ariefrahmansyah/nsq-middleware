# peacock [![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/ariefrahmansyah/peacock) [![CircleCI](https://circleci.com/gh/ariefrahmansyah/peacock/tree/master.png?style=shield)](https://circleci.com/gh/ariefrahmansyah/peacock/tree/master) [![Coverage Status](https://coveralls.io/repos/github/ariefrahmansyah/peacock/badge.svg?branch=master)](https://coveralls.io/github/ariefrahmansyah/peacock?branch=master) [![GoReportCard](https://goreportcard.com/badge/github.com/ariefrahmansyah/peacock)](https://goreportcard.com/report/github.com/ariefrahmansyah/peacock)

Peacock is middleware for your NSQ consumer handler function. It's heavily inspired by negroni.

![Peacock](https://www.web-savvy-marketing.com/wp-content/uploads/2013/11/Be-the-Peacock.jpg)
Image source: https://www.web-savvy-marketing.com/2013/11/peacock/

## Usage
We can create new middleware object that implement nsq.Handler interface and use it in Peacock object using Use method.

We also can use nsq.HandlerFunc that we already created before.

```go
// Create peacock object.
peacock := peacock.New()

// Create middleware object that implement nsq.Handler interface.
type Middleware1 struct{}
func (m1 Middleware1) HandleMessage(message *nsq.Message, next nsq.HandlerFunc) error {
	log.Printf("Middleware 1:\t%s\n", message.Body)
	return next(message)
}

// Use Middleware1.
peacock.Use(Middleware1{})

// We may already have nsq.HandlerFunc.
func handlerFunc1(message *nsq.Message) error {
	log.Printf("Handler Func 1:\t%s\n\n", message.Body)
	message.Finish()
	return nil
}

// Let's use it too.
peacock.UseHandlerFunc(handlerFunc1)

consumer, _ := nsq.NewConsumer(topicName, channelName, nsq.NewConfig())
consumer.AddHandler(peacock)
consumer.ConnectToNSQD(nsqdAddress)
```
