package nsqmiddleware

import (
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dflBuckets = []float64{300, 1000, 2500, 5000}
)

const (
	messageName  = "consumer_messages_total"
	durationName = "consumer_duration_milliseconds"
)

// PromMiddlewareOpts specifies options how to create new PromMiddleware.
type PromMiddlewareOpts struct {
	// Buckets specifies an custom buckets to be used in request histograpm.
	Buckets []float64
}

// PromMiddleware is a handler that exposes prometheus metrics
// for the number of messages, and the process duration,
// partitioned by topic, channel, attempt and status.
type PromMiddleware struct {
	message *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func NewPromMiddleware(namespace string, opt PromMiddlewareOpts) *PromMiddleware {
	var pm PromMiddleware

	pm.message = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      messageName,
			Help:      "How many NSQ messages processed, partitioned by topic, channel, attempt and status.",
		},
		[]string{"topic", "channe", "attempt", "status"},
	)
	prometheus.MustRegister(pm.message)

	buckets := opt.Buckets
	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	pm.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      durationName,
		Help:      "How long it took to consume the message, partitioned by topic, channel, attempt and status.",
		Buckets:   buckets,
	},
		[]string{"topic", "channe", "attempt", "status"},
	)
	prometheus.MustRegister(pm.latency)

	return &pm
}

func (promM PromMiddleware) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	start := time.Now()
	status := "ok"

	err := next(message)
	if err != nil {
		status = "error"
	}

	go promM.message.WithLabelValues(topic, channel, fmt.Sprint(message.Attempts), status).Inc()
	go promM.latency.WithLabelValues(topic, channel, fmt.Sprint(message.Attempts), status).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)

	return err
}
