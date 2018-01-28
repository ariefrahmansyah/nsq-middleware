package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	nsqm "github.com/ariefrahmansyah/nsq-middleware"
	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
)

var nsqd = "127.0.0.1:4150"
var topic = "ar_simple_test"
var channel = "ar_simple_channel"
var maxRandom = 23031994

func main() {
	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)

	producer, err := nsq.NewProducer(nsqd, nsq.NewConfig())
	if err != nil {
		log.Fatalln(err)
	}

	// Produce message
	go func() {
		rand.Seed(time.Now().UnixNano())
		for {
			randNumber := rand.Intn(maxRandom)
			data := struct {
				RandomNumber int `json:"random_number"`
			}{randNumber}

			dataJSON, _ := json.Marshal(data)

			producer.Publish(topic, dataJSON)
			log.Printf("Published:\t\t%s\n", dataJSON)

			time.Sleep(time.Second)
		}
	}()

	middleware1 := Middleware1{}

	var handler1 nsq.HandlerFunc
	handler1 = func(message *nsq.Message) error {
		log.Printf("Handler 1:\t\t%s\n", message.Body)
		return nil
	}

	nsqMid := nsqm.New(topic, channel)
	nsqMid.Use(nsqm.NewPromMiddleware("test", nsqm.PromMiddlewareOpts{}))
	nsqMid.Use(middleware1)
	nsqMid.UseHandler(handler1)
	nsqMid.UseHandlerFunc(handlerFunc1)

	consumer, err := nsq.NewConsumer(topic, channel, nsq.NewConfig())
	if err != nil {
		log.Fatalln(err)
	}

	consumer.AddHandler(nsqMid)

	consumer.ConnectToNSQD(nsqd)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/metrics", prometheus.Handler())

	http.HandleFunc("/ping", ping)
	http.ListenAndServe(":"+port, nil)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

type Middleware1 struct{}

func (m1 Middleware1) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	log.Printf("Middleware 1:\t%s -> %s\t%s\n", topic, channel, message.Body)
	return next(message)
}

func handlerFunc1(message *nsq.Message) error {
	log.Printf("Handler Func 1:\t%s\n\n", message.Body)
	message.Finish()
	return nil
}

func waitForTerminate(ctx context.Context, cancel context.CancelFunc) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	defer func() {
		signal.Stop(gracefulStop)
		cancel()
	}()

	select {
	case <-gracefulStop:
		cancel()
	case <-ctx.Done():
	}
}
