package nsqmiddleware

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/alecthomas/template"
	"github.com/nsqio/go-nsq"
)

// Level type.
type Level uint32

// These are the different logging levels.
const (
	SuccessLevel Level = iota
	ErrorLevel
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

// LoggerDefaultLevel is the log level used by the default Logger instance.
var LoggerDefaultLevel = SuccessLevel

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
	level      Level
}

// NewLogger returns a new Logger instance.
func NewLogger() *Logger {
	logger := &Logger{ILogger: log.New(os.Stdout, "[nsqm] ", 0), dateFormat: LoggerDefaultDateFormat}
	logger.SetLevel(LoggerDefaultLevel)
	logger.SetFormat(LoggerDefaultFormat)
	return logger
}

// SetLevel sets the level.
func (logger *Logger) SetLevel(level Level) {
	logger.level = level
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
	validToLog := true

	err := next(message)
	if err == nil && logger.level >= ErrorLevel {
		validToLog = false
	} else if err != nil {
		status = "error"
		errStr = err.Error()
	}

	if validToLog {
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
	}

	return err
}
