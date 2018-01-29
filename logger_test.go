package nsqmiddleware

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
)

func TestLoggerMiddleware(t *testing.T) {
	var buff bytes.Buffer

	logger := NewLogger()
	logger.SetDateFormat(time.RFC3339)
	logger.ILogger = log.New(&buff, "[nsqm] ", 0)

	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(logger)
	nsqMid.Use(mockMiddleware{})
	nsqMid.HandleMessage(&nsq.Message{Attempts: 1, Body: []byte(`{"message": 1}`)})

	if !strings.Contains(buff.String(), "nsqm") {
		t.Errorf("log does not contain nsqm tag")
	}

	if len(buff.String()) == 0 {
		t.Errorf("log body must not empty 😱")
	}
}

func TestLoggerMiddlewareError(t *testing.T) {
	var buff bytes.Buffer

	logger := NewLogger()
	logger.SetDateFormat(time.RFC3339)
	logger.ILogger = log.New(&buff, "[nsqm] ", 0)

	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(logger)
	nsqMid.UseHandlerFunc(nsqHandlerFuncError)
	nsqMid.HandleMessage(&nsq.Message{Attempts: 1, Body: []byte(`{"message": 1}`)})

	if !strings.Contains(buff.String(), "error") {
		t.Errorf("log does not contain error")
	}

	if len(buff.String()) == 0 {
		t.Errorf("log body must not empty 😱")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buff bytes.Buffer

	logger := NewLogger()
	logger.SetLevel(ErrorLevel)
	logger.ILogger = log.New(&buff, "[nsqm] ", 0)

	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(logger)
	nsqMid.Use(mockMiddleware{})
	nsqMid.HandleMessage(&nsq.Message{Attempts: 1, Body: []byte(`{"message": 1}`)})

	if len(buff.String()) != 0 {
		t.Errorf("log body must be empty 😱")
	}
}

func TestLogger_SetFormat(t *testing.T) {
	var buff bytes.Buffer

	logger := NewLogger()
	logger.SetFormat("{{.Attempts}}")
	logger.ILogger = log.New(&buff, "[nsqm] ", 0)

	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(logger)
	nsqMid.Use(mockMiddleware{})
	nsqMid.HandleMessage(&nsq.Message{Attempts: 1, Body: []byte(`{"message": 1}`)})

	if !strings.Contains(buff.String(), "[nsqm] 1") {
		t.Errorf("expected log output is wrong. got: %s", buff.String())
	}

	if len(buff.String()) == 0 {
		t.Errorf("log body must not empty 😱")
	}
}
