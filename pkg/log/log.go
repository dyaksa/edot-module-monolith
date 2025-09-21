package log

import (
	"time"
)

type Logger interface {
	Info(msg string, fn ...LoggerContextFn)
	Error(msg string, fn ...LoggerContextFn)
	Warn(msg string, fn ...LoggerContextFn)
	Debug(msg string, fn ...LoggerContextFn)
	Fatal(msg string, fn ...LoggerContextFn)
	Panic(msg string, fn ...LoggerContextFn)
}

type LoggerContextFn func(LoggerContext)

type LoggerContext interface {
	Any(key string, value any)
	Bool(key string, value bool)
	Bytes(key string, value []byte)
	String(key string, value string)
	Float64(key string, value float64)
	Int64(key string, value int64)
	Uint64(key string, value uint64)
	Time(key string, value time.Time)
	Duration(key string, value time.Duration)
	Error(key string, err error)
}

type Loggable interface {
	AsLog() any
}
