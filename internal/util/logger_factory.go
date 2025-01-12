package util

type LoggerFactory interface {
	NewLogger(name string) Logger
}

type loggerFactory struct{}

func NewLoggerFactory() LoggerFactory {
	return &loggerFactory{}
}

func (f *loggerFactory) NewLogger(name string) Logger {
	return NewLogger(name)
}
