package logrus

import (
	"fmt"
	"time"

	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/sirupsen/logrus"
)

type Level int

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Opts func(*logrusLogger) error

func withLevelString(level string) int {
	switch level {
	case "panic":
		return int(PanicLevel)
	case "fatal":
		return int(FatalLevel)
	case "error":
		return int(ErrorLevel)
	case "warn", "warning":
		return int(WarnLevel)
	case "info":
		return int(InfoLevel)
	case "debug":
		return int(DebugLevel)
	case "trace":
		return int(TraceLevel)
	}
	return int(DebugLevel)
}

func WithLevel(lvl string) Opts {
	return func(ll *logrusLogger) error {
		level := withLevelString(lvl)
		if Level(level) > Level(logrus.DebugLevel) || Level(level) < Level(logrus.PanicLevel) {
			return fmt.Errorf("invalid log level")
		}

		ll.level = Level(level)
		return nil
	}
}

func WithJSONFormatter() Opts {
	return func(ll *logrusLogger) error {
		ll.logrus.SetFormatter(&logrus.JSONFormatter{})
		return nil
	}
}

func WithCaller(status bool) Opts {
	return func(ll *logrusLogger) error {
		ll.caller = status
		return nil
	}
}

type logrusLogger struct {
	logrus *logrus.Logger
	caller bool
	level  Level
}

type loggerContext struct {
	fields logrus.Fields
}

func New(opts ...Opts) (*logrusLogger, error) {
	var log = &logrusLogger{
		logrus: logrus.New(),
	}

	for _, opt := range opts {
		if err := opt(log); err != nil {
			return nil, err
		}
	}

	log.logrus.SetLevel(logrus.Level(log.level))
	log.logrus.SetReportCaller(log.caller)

	return log, nil
}

func (l *logrusLogger) Info(msg string, fn ...log.LoggerContextFn) {
	if l.level > InfoLevel {
		return
	}

	l.logrus.WithFields(newLoggerContext(fn...).fields).Info(msg)
}

func (l *logrusLogger) Error(msg string, fn ...log.LoggerContextFn) {
	if l.level > ErrorLevel {
		return
	}

	l.logrus.WithFields(newLoggerContext(fn...).fields).Error(msg)
}

func (l *logrusLogger) Warn(msg string, fn ...log.LoggerContextFn) {
	if l.level > WarnLevel {
		return
	}

	l.logrus.WithFields(newLoggerContext(fn...).fields).Warn(msg)
}

func (l *logrusLogger) Debug(msg string, fn ...log.LoggerContextFn) {
	if l.level > DebugLevel {
		return
	}

	l.logrus.WithFields(newLoggerContext(fn...).fields).Debug(msg)
}

func (l *logrusLogger) Fatal(msg string, fn ...log.LoggerContextFn) {
	if l.level > FatalLevel {
		return
	}

	l.logrus.WithFields(newLoggerContext(fn...).fields).Fatal(msg)
}

func (l *logrusLogger) Panic(msg string, fn ...log.LoggerContextFn) {
	if l.level > PanicLevel {
		return
	}

	l.logrus.WithFields(newLoggerContext(fn...).fields).Panic(msg)
}

func newLoggerContext(fn ...log.LoggerContextFn) *loggerContext {
	ctx := &loggerContext{
		fields: logrus.Fields{},
	}

	for _, f := range fn {
		f(ctx)
	}
	return ctx
}

func (lc *loggerContext) Any(key string, value interface{}) {
	lc.fields[key] = value
}

func (lc *loggerContext) Bool(key string, value bool) {
	lc.fields[key] = value
}

func (lc *loggerContext) Bytes(key string, value []byte) {
	lc.fields[key] = value
}

func (lc *loggerContext) String(key string, value string) {
	lc.fields[key] = value
}

func (lc *loggerContext) Float64(key string, value float64) {
	lc.fields[key] = value
}

func (lc *loggerContext) Int64(key string, value int64) {
	lc.fields[key] = value
}

func (lc *loggerContext) Uint64(key string, value uint64) {
	lc.fields[key] = value
}

func (lc *loggerContext) Time(key string, value time.Time) {
	lc.fields[key] = value
}

func (lc *loggerContext) Duration(key string, value time.Duration) {
	lc.fields[key] = value
}

func (lc *loggerContext) Error(key string, err error) {
	lc.fields[key] = err
}
