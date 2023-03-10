package logger

type Logger interface {
	Init(*Config) (Logger, error)
	NoCaller() Logger

	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
