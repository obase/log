package log

import (
	"context"
)

/*日志适配方法*/
type LoggerFunc func(ctx context.Context, format string, args ...interface{})

// 全局日志输出适配, 应用可以写日志前覆盖此值
var (
	Errorf LoggerFunc = Error
	Inforf LoggerFunc = Info
	Warnf  LoggerFunc = Warn
	Debugf LoggerFunc = Debug
	Flushf func()     = Flush
)
