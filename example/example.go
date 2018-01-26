package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ariefrahmansyah/peacock"
	"github.com/nsqio/go-nsq"
)

var nsqd = "127.0.0.1:4150"
var topic = "ar_simple_test"
var channel = "ar_simple_channel"
var maxRandom = 23031994

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

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

	peacock := peacock.New()
	peacock.Use(middleware1)
	peacock.UseHandlerFunc(consumeHandler2)

	consumer, err := nsq.NewConsumer(topic, channel, nsq.NewConfig())
	if err != nil {
		log.Fatalln(err)
	}

	consumer.AddHandler(peacock)

	consumer.ConnectToNSQD(nsqd)

	waitForTerminate(ctx, cancel)
}

type Middleware1 struct{}

func (m1 Middleware1) HandleMessage(msg *nsq.Message, next nsq.HandlerFunc) error {
	log.Printf("Middleware 1:\t%s\n", msg.Body)
	return next(msg)
}

func consumeHandler2(msg *nsq.Message) error {
	log.Printf("Consumer 2:\t\t%s\n", msg.Body)
	msg.Finish()
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
