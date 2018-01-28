package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
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

	var handler1 nsq.HandlerFunc
	handler1 = func(message *nsq.Message) error {
		log.Printf("Handler 1:\t\t%s\n", message.Body)
		return nil
	}

	for i := 1; i <= 3; i++ {
		channelName := fmt.Sprintf("%s_%d", channel, i)

		nsqMid := nsqm.New(topic, channelName)
		nsqMid.Use(nsqm.NewRecovery())
		nsqMid.Use(nsqm.NewLogger())
		nsqMid.Use(nsqm.NewPrometheus())
		nsqMid.Use(Middleware1{})
		nsqMid.UseHandler(handler1)
		nsqMid.UseHandlerFunc(handlerFunc1)

		consumer, err := nsq.NewConsumer(topic, channelName, nsq.NewConfig())
		if err != nil {
			log.Fatalln(err)
		}

		consumer.AddHandler(nsqMid)

		consumer.ConnectToNSQD(nsqd)
	}

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
	log.Printf("Handler Func 1:\t%s\n", message.Body)
	message.Finish()
	return nil
}
