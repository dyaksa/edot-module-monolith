package log

import "time"

func Any(key string, value any) LoggerContextFn {
	if l, ok := value.(Loggable); ok {
		value = l.AsLog()
	}

	return func(ctx LoggerContext) {
		ctx.Any(key, value)
	}

}

func Bool(key string, value bool) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Bool(key, value)
	}
}

func Bytes(key string, value []byte) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Bytes(key, value)
	}
}

func String(key string, value string) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.String(key, value)
	}
}

func Float64(key string, value float64) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Float64(key, value)
	}
}

func Int64(key string, value int64) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Int64(key, value)
	}
}

func Uint64(key string, value uint64) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Uint64(key, value)
	}
}

func Time(key string, value time.Time) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Time(key, value)
	}
}

func Duration(key string, value time.Duration) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Duration(key, value)
	}
}

func Error(key string, err error) LoggerContextFn {
	return func(lc LoggerContext) {
		lc.Error(key, err)
	}
}
