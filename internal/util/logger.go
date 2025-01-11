package util

import (
	"context"
	"github.com/rs/zerolog"
	"os"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	WithContext(ctx context.Context) Logger
}

type logger struct {
	log         zerolog.Logger
	ctx         context.Context
	serviceName string
}

func NewLogger(serviceName string) Logger {
	// Configure console writer with colors
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02T15:04:05.000Z07:00",
		NoColor:    false,
	}

	// Create logger
	log := zerolog.New(output).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Str("service", serviceName).
		Caller().
		Logger()

	return &logger{
		log:         log,
		serviceName: serviceName,
	}
}

func (l *logger) WithContext(ctx context.Context) Logger {
	return &logger{
		log:         l.log,
		ctx:         ctx,
		serviceName: l.serviceName,
	}
}

func (l *logger) Info(msg string, args ...any) {
	logEvent := l.log.Info()
	addFields(logEvent, args...)
	logEvent.Msg(msg)
}

func (l *logger) Error(msg string, args ...any) {
	logEvent := l.log.Error()
	addFields(logEvent, args...)
	logEvent.Msg(msg)
}

func (l *logger) Debug(msg string, args ...any) {
	logEvent := l.log.Debug()
	addFields(logEvent, args...)
	logEvent.Msg(msg)
}

func (l *logger) Warn(msg string, args ...any) {
	logEvent := l.log.Warn()
	addFields(logEvent, args...)
	logEvent.Msg(msg)
}

// addFields adds key-value pairs to the log event
func addFields(event *zerolog.Event, args ...any) {
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				event.Interface(key, args[i+1])
			}
		}
	}
}
