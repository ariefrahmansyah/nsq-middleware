package nsqmiddleware

import (
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	promMessageName  = "nsqm_consumer_messages_total"
	promDurationName = "nsqm_consumer_duration_milliseconds"
)

var (
	promMessage *prometheus.CounterVec
	promLatency *prometheus.HistogramVec
)

func init() {
	promMessage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: promMessageName,
			Help: "How many NSQ messages processed, partitioned by topic, channel, attempts and status.",
		},
		[]string{"topic", "channel", "attempts", "status"},
	)
	prometheus.MustRegister(promMessage)

	buckets := []float64{300, 1000, 2500, 5000}
	promLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    promDurationName,
		Help:    "How long it took to consume the message, partitioned by topic, channel, attempts and status.",
		Buckets: buckets,
	},
		[]string{"topic", "channel", "attempts", "status"},
	)
	prometheus.MustRegister(promLatency)
}

// Prometheus is a handler that exposes prometheus metrics
// for the number of messages, and the process duration,
// partitioned by topic, channel, attempts and status.
type Prometheus struct{}

// NewPrometheus returns a new Prometheus Middleware instance.
func NewPrometheus() *Prometheus {
	return &Prometheus{}
}

func (prometheus Prometheus) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	start := time.Now()
	status := "ok"

	err := next(message)
	if err != nil {
		status = "error"
	}

	go promMessage.WithLabelValues(topic, channel, fmt.Sprint(message.Attempts), status).Inc()
	go promLatency.WithLabelValues(topic, channel, fmt.Sprint(message.Attempts), status).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)

	return err
}
