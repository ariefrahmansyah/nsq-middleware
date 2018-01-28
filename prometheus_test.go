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

	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(NewPromMiddleware("test", PromMiddlewareOpts{}))
	nsqMid.Use(mockMiddleware{})

	nsqMid.HandleMessage(&nsq.Message{Body: []byte(`{"message": 1}`)})

	reqMetrics, err := http.NewRequest("GET", "http://localhost:8080/metrics", nil)
	if err != nil {
		t.Error(err)
	}

	n.ServeHTTP(recorder, reqMetrics)

	body := recorder.Body.String()

	if !strings.Contains(body, messageName) {
		t.Errorf("body does not contain total consumed messages '%s'", messageName)
	}

	if !strings.Contains(body, durationName) {
		t.Errorf("body does not contain consumer duration '%s'", durationName)
	}
}
