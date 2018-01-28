package nsqmiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nsqio/go-nsq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/negroni"
)

func TestPrometheusMiddleware(t *testing.T) {
	recorder := httptest.NewRecorder()

	n := negroni.New()
	r := http.NewServeMux()
	r.Handle("/metrics", prometheus.Handler())
	n.UseHandler(r)

	// Success handler
	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(NewPrometheus())
	nsqMid.Use(mockMiddleware{})
	nsqMid.HandleMessage(&nsq.Message{Body: []byte(`{"message": 1}`)})

	// Error handler
	nsqMid = New(defaultTopic, defaultChannel)
	nsqMid.Use(NewPrometheus())
	nsqMid.UseHandlerFunc(nsqHandlerFuncError)
	nsqMid.HandleMessage(&nsq.Message{Body: []byte(`{"message": 1}`)})

	reqMetrics, err := http.NewRequest("GET", "http://localhost:8080/metrics", nil)
	if err != nil {
		t.Error(err)
	}

	n.ServeHTTP(recorder, reqMetrics)

	body := recorder.Body.String()

	if !strings.Contains(body, promMessageName) && !strings.Contains(body, "ok") && !strings.Contains(body, "error") {
		t.Errorf("body does not contain all expected consumed messages '%s' metrics!", promMessageName)
	}

	if !strings.Contains(body, promDurationName) {
		t.Errorf("body does not contain consumer duration '%s'", promDurationName)
	}
}
