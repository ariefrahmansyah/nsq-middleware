package nsqmiddleware

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/alecthomas/template"
	"github.com/nsqio/go-nsq"
)

// LoggerEntry is the structure passed to the template.
type LoggerEntry struct {
	StartTime   string
	Status      string
	Duration    time.Duration
	Topic       string
	Channel     string
	Attempts    uint16
	ErrorString string
}

// LoggerDefaultFormat is the format logged used by the default Logger instance.
var LoggerDefaultFormat = "{{.StartTime}} \t{{.Duration}} [{{.Topic}}/{{.Channel}}] ({{.Attempts}}) {{.Status}} {{.ErrorString}} \n"

// LoggerDefaultDateFormat is the format used for date by the default Logger instance.
var LoggerDefaultDateFormat = time.RFC3339

// ILogger interface
type ILogger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Logger is a middleware handler that logs the messages as it goes in and the error as it goes out.
type Logger struct {
	// ILogger implements just enough log.Logger interface to be compatible with other implementations
	ILogger
	dateFormat string
	template   *template.Template
}

// NewLogger returns a new Logger instance.
func NewLogger() *Logger {
	logger := &Logger{ILogger: log.New(os.Stdout, "[nsqm] ", 0), dateFormat: LoggerDefaultDateFormat}
	logger.SetFormat(LoggerDefaultFormat)
	return logger
}

// SetFormat sets the format used by the logger.
func (logger *Logger) SetFormat(format string) {
	logger.template = template.Must(template.New("nsqm_parser").Parse(format))
}

// SetDateFormat sets the format used for date by the logger.
func (logger *Logger) SetDateFormat(format string) {
	logger.dateFormat = format
}

func (logger *Logger) HandleMessage(topic, channel string, message *nsq.Message, next nsq.HandlerFunc) error {
	start := time.Now()
	status := "ok"
	errStr := ""

	err := next(message)
	if err != nil {
		status = "error"
		errStr = err.Error()
	}

	log := LoggerEntry{
		StartTime:   start.Format(logger.dateFormat),
		Status:      status,
		Duration:    time.Since(start),
		Topic:       topic,
		Channel:     channel,
		Attempts:    message.Attempts,
		ErrorString: errStr,
	}

	buff := &bytes.Buffer{}
	logger.template.Execute(buff, log)
	logger.Printf(buff.String())

	return err
}
