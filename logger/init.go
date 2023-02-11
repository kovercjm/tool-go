package logger

import (
	"os"
)

type newLogger struct {
	Logger
}

func New(config *Config, options ...Option) (Logger, error) {
	l := &newLogger{}
	for _, option := range options {
		option(l)
	}
	if l.Logger == nil {
		l.Logger = ZapLogger{}
	}

	return l.Logger.Init(config)
}

func Default() (Logger, error) {
	deployment := os.Getenv("DEPLOYMENT") // try to get deployment name from env
	if deployment == "" {
		deployment = "default"
	}
	return New(&Config{Deployment: deployment})
}
