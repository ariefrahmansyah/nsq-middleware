package nsqmiddleware

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/nsqio/go-nsq"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	handlerFunc = func(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
		log.Printf("handlerFunc called. Topic: %s. Channel: %s. Message Body: %s", topic, channel, message.Body)
		return next(message)
	}

	nsqHandlerFunc = func(message *nsq.Message) error {
		log.Printf("nsqHandlerFunc called. Message Body: %s", message.Body)
		return nil
	}

	nsqHandlerFuncSuccess = nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("nsqHandlerFuncSuccess called. Message Body: %s", message.Body)
		return nil
	})

	nsqHandlerFuncError = nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("nsqHandlerFuncError called. Message Body: %s", message.Body)
		return errors.New("error")
	})

	nsqHandlerFuncPanic = nsq.HandlerFunc(func(message *nsq.Message) error {
		panic("panic at the disco 👨‍🎤")
		return nil
	})

	run := m.Run()
	return run
}

type mockMiddleware struct {
	err error
}

func (m mockMiddleware) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	log.Printf("Mock middleware. Topic: %s. Channel: %s. Message Body: %s", topic, channel, message.Body)

	if m.err != nil {
		log.Printf("Mock middleware error: %s", m.err)
		return m.err
	}

	return next(message)
}

var handlerFunc HandlerFunc

var nsqHandlerFunc nsq.HandlerFunc

var nsqHandlerFuncSuccess nsq.HandlerFunc
var nsqHandlerFuncError nsq.HandlerFunc
var nsqHandlerFuncPanic nsq.HandlerFunc

var defaultTopic = "topic_test"
var defaultChannel = "channel_test"

func TestWrapHandler(t *testing.T) {
	got := WrapHandler(nsqHandlerFuncSuccess)
	if got == nil {
		t.Errorf("WrapHandler() must not nil")
	}
}

func TestHandlerFunc_HandleMessage(t *testing.T) {
	type args struct {
		topic   string
		channel string
		message *nsq.Message
		next    nsq.HandlerFunc
	}
	tests := []struct {
		name        string
		handlerFunc HandlerFunc
		args        args
		wantErr     bool
	}{
		{
			"success handler",
			handlerFunc,
			args{
				"topic_1",
				"channel_1",
				&nsq.Message{},
				nsqHandlerFuncSuccess,
			},
			false,
		},
		{
			"error handler",
			handlerFunc,
			args{
				"topic_1",
				"channel_1",
				&nsq.Message{},
				nsqHandlerFuncError,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.handlerFunc.HandleMessage(tt.args.topic, tt.args.channel, tt.args.message, tt.args.next); (err != nil) != tt.wantErr {
				t.Errorf("HandlerFunc.HandleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMiddleware_HandleMessage(t *testing.T) {
	type fields struct {
		handler Handler
		next    *middleware
	}
	type args struct {
		message *nsq.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := middleware{
				handler: tt.fields.handler,
				next:    tt.fields.next,
			}
			if err := middleware.HandleMessage(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("middleware.HandleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_buildMiddleware(t *testing.T) {
	type args struct {
		topic    string
		channel  string
		handlers []Handler
	}
	tests := []struct {
		name string
		args args
		want middleware
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMiddleware(tt.args.topic, tt.args.channel, tt.args.handlers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_emptyMiddleware(t *testing.T) {
	tests := []struct {
		name string
		want middleware
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emptyMiddleware(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("emptyMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		topic    string
		channel  string
		handlers []Handler
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"empty handler",
			args{},
		},
		{
			"one handler",
			args{
				"topic_1",
				"channel_1",
				[]Handler{
					handlerFunc,
				},
			},
		},
		{
			"two handler",
			args{
				"topic_1",
				"channel_1",
				[]Handler{
					handlerFunc,
					handlerFunc,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.topic, tt.args.channel, tt.args.handlers...)
			if got == nil {
				t.Errorf("New() must not nil")
			}
		})
	}
}

func TestNSQM_HandleMessage(t *testing.T) {
	type fields struct {
		handlers   []Handler
		middleware middleware
	}
	type args struct {
		message *nsq.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"empty handler",
			fields{
				[]Handler{},
				buildMiddleware("", "", []Handler{}),
			},
			args{
				&nsq.Message{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nsqm := NSQM{
				handlers:   tt.fields.handlers,
				middleware: tt.fields.middleware,
			}
			if err := nsqm.HandleMessage(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("NSQM.HandleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNSQM_Use(t *testing.T) {
	type fields struct {
		handlers   []Handler
		middleware middleware
	}
	type args struct {
		handler Handler
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPanic bool
	}{
		{
			"nil handler",
			fields{
				[]Handler{},
				middleware{},
			},
			args{
				nil,
			},
			true,
		},
		{
			"1",
			fields{
				[]Handler{},
				middleware{},
			},
			args{
				mockMiddleware{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if tt.wantPanic {
					if r := recover(); r != nil {
						err, ok := r.(error)
						if !ok {
							err = fmt.Errorf("pkg: %v", r)
						}
						log.Printf("panic recovered ( %s ). { %s }", tt.name, err)
					}
				}
			}()

			nsqm := &NSQM{
				handlers:   tt.fields.handlers,
				middleware: tt.fields.middleware,
			}
			nsqm.Use(tt.args.handler)
		})
	}
}

func TestNSQM_UseFunc(t *testing.T) {
	type fields struct {
		handlers   []Handler
		middleware middleware
	}
	type args struct {
		handlerFunc func(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"1",
			fields{
				[]Handler{},
				middleware{},
			},
			args{
				handlerFunc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nsqm := &NSQM{
				handlers:   tt.fields.handlers,
				middleware: tt.fields.middleware,
			}
			nsqm.UseFunc(tt.args.handlerFunc)
		})
	}
}

func TestNSQM_UseHandler(t *testing.T) {
	type fields struct {
		handlers   []Handler
		middleware middleware
	}
	type args struct {
		handler nsq.Handler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"1",
			fields{
				[]Handler{},
				middleware{},
			},
			args{
				nsqHandlerFunc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nsqm := &NSQM{
				handlers:   tt.fields.handlers,
				middleware: tt.fields.middleware,
			}
			nsqm.UseHandler(tt.args.handler)
		})
	}
}

func TestNSQM_UseHandlerFunc(t *testing.T) {
	type fields struct {
		handlers   []Handler
		middleware middleware
	}
	type args struct {
		handlerFunc func(message *nsq.Message) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"1",
			fields{
				[]Handler{},
				middleware{},
			},
			args{
				nsqHandlerFuncSuccess,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nsqm := &NSQM{
				handlers:   tt.fields.handlers,
				middleware: tt.fields.middleware,
			}
			nsqm.UseHandlerFunc(tt.args.handlerFunc)
		})
	}
}
