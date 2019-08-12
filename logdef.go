package log

import (
	"context"
)

// 中间定义,可以覆盖
var (
	Debug func(ctx context.Context, format string, args ...interface{})
	Info  func(ctx context.Context, format string, args ...interface{})
	Warn  func(ctx context.Context, format string, args ...interface{})
	Error func(ctx context.Context, format string, args ...interface{})
	Fatal func(ctx context.Context, format string, args ...interface{})
	Flush func()
)

// 历史API
var (
	// Deprecated: please use Info instead
	Debugf func(ctx context.Context, format string, args ...interface{})
	// Deprecated: please use Info instead
	Infof func(ctx context.Context, format string, args ...interface{})
	// Deprecated: please use Info instead
	Inforf func(ctx context.Context, format string, args ...interface{})
	// Deprecated: please use Info instead
	Warnf func(ctx context.Context, format string, args ...interface{})
	// Deprecated: please use Info instead
	Errorf func(ctx context.Context, format string, args ...interface{})
	// Deprecated: please use Info instead
	Fatalf func(ctx context.Context, format string, args ...interface{})
	// Deprecated: please use Info instead
	Flushf func()
)
