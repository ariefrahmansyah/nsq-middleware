package nsqmiddleware

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/nsqio/go-nsq"
)

func TestRecoveryMiddleware(t *testing.T) {
	buff := bytes.NewBufferString("")

	recovery := NewRecovery()
	recovery.Logger = log.New(buff, "[nsqm] ", 0)

	nsqMid := New(defaultTopic, defaultChannel)
	nsqMid.Use(recovery)
	nsqMid.UseHandlerFunc(nsqHandlerFuncPanic)
	nsqMid.HandleMessage(&nsq.Message{Attempts: 1, Body: []byte(`{"message": 1}`)})

	if !strings.Contains(buff.String(), "PANIC") {
		t.Errorf("log does not contain PANIC")
	}

	if len(buff.String()) == 0 {
		t.Errorf("log body must not empty 😱")
	}
}
