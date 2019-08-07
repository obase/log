package log

import (
	"context"
)

/*日志适配方法*/
var Fatal func(ctx context.Context, format string, args ...interface{})
var Error func(ctx context.Context, format string, args ...interface{})
var Warn func(ctx context.Context, format string, args ...interface{})
var Info func(ctx context.Context, format string, args ...interface{})
var Debug func(ctx context.Context, format string, args ...interface{})
var Flush func() = func() {
	if _default != nil {
		_default.Flush()
	}
	for _, v := range _loggers {
		v.Flush()
	}
}

/*兼容旧的API*/
// Deprecated: please use Fatal instead
var Fatalf func(ctx context.Context, format string, args ...interface{}) = Fatal

// Deprecated: please use Error instead
var Errorf func(ctx context.Context, format string, args ...interface{}) = Error

// Deprecated: please use Warn instead
var Warnf func(ctx context.Context, format string, args ...interface{}) = Warn

// Deprecated: please use Info instead
var Infof func(ctx context.Context, format string, args ...interface{}) = Info

// Deprecated: please use Info instead
var Inforf func(ctx context.Context, format string, args ...interface{}) = Info

// Deprecated: please use Debug instead
var Debugf func(ctx context.Context, format string, args ...interface{}) = Debug

// Deprecated: please use Flush instead
var Flushf func() = Flush
