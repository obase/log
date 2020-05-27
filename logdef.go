/*
实现log代理层接口功能,默认使用builtin的logger实现. 但开发可以调用Setup()重置内置的日志逻辑. 但不保证并发安全.
*/
package log

import (
	"context"
	"fmt"
	"os"
	"time"
)

const DEFAULT_FLUSH_PERIOD = 5 * time.Second

type Level uint

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	Level Level
	Log   func(level Level, args ...interface{})
	Logf  func(level Level, format string, args ...interface{})
	Flush func()
	Close func()
}

func (lg *Logger) Debug(args ...interface{}) {
	if DEBUG >= lg.Level {
		lg.Log(DEBUG, args...)
	}
}

func (lg *Logger) Debugf(format string, args ...interface{}) {
	if DEBUG >= lg.Level {
		lg.Logf(DEBUG, format, args...)
	}
}

func (lg *Logger) Info(args ...interface{}) {
	if INFO >= lg.Level {
		lg.Log(INFO, args...)
	}
}

func (lg *Logger) Infof(format string, args ...interface{}) {
	if INFO >= lg.Level {
		lg.Logf(INFO, format, args...)
	}
}

func (lg *Logger) Warn(args ...interface{}) {
	if WARN >= lg.Level {
		lg.Log(WARN, args...)
	}
}

func (lg *Logger) Warnf(format string, args ...interface{}) {
	if WARN >= lg.Level {
		lg.Logf(WARN, format, args...)
	}
}

func (lg *Logger) Error(args ...interface{}) {
	if ERROR >= lg.Level {
		lg.Log(ERROR, args...)
	}
}

func (lg *Logger) Errorf(format string, args ...interface{}) {
	if ERROR >= lg.Level {
		lg.Logf(ERROR, format, args...)
	}
}

func (lg *Logger) ErrorStack(format string, args ...interface{}) {
	if ERROR >= lg.Level {
		format = format + "\n%s"
		args = append(args, stack(false))
		lg.Logf(ERROR, format, args...)
	}
}

func (lg *Logger) Fatal(args ...interface{}) {
	if FATAL >= lg.Level {
		lg.Log(FATAL, args...)
		Close() // FATAL前关闭掉所有日志,避免损失丢失关键信息
		os.Exit(FATAL_EXIT_CODE)
	}
}

func (lg *Logger) Fatalf(format string, args ...interface{}) {
	if FATAL >= lg.Level {
		lg.Logf(FATAL, format, args...)
		Close() // FATAL前关闭掉所有日志,避免损失丢失关键信息
		os.Exit(FATAL_EXIT_CODE)
	}
}

func (lg *Logger) FatalStack(format string, args ...interface{}) {
	if FATAL >= lg.Level {
		format = format + "\n%s"
		args = append(args, stack(false))
		lg.Logf(FATAL, format, args...)
		Close() // FATAL前关闭掉所有日志,避免损失丢失关键信息
		os.Exit(FATAL_EXIT_CODE)
	}
}

var (
	glog *Logger
	gmap = make(map[string]*Logger)
	gctx context.Context
	gcnf context.CancelFunc
)

// 用于替换
func Setup(flushPeriod time.Duration, g *Logger, m map[string]*Logger) {

	// 先关闭已经打开日志句柄与刷新线程
	Close()

	glog = g
	for k, v := range m {
		gmap[k] = v
	}

	// 如果未指定刷新周期则则用默认值
	if flushPeriod <= 0 {
		flushPeriod = DEFAULT_FLUSH_PERIOD
	}
	gctx, gcnf = context.WithCancel(context.Background())
	go flush(gctx, flushPeriod, g, m)

}

func Get(name string) (ret *Logger) {
	if name == "" {
		ret = glog
	}
	ret = gmap[name]
	return
}

func Must(name string) (ret *Logger) {
	if name == "" {
		ret = glog
	}
	ret = gmap[name]
	if ret == nil {
		panic("invalid logger " + name)
	}
	return
}

func Debug(args ...interface{}) {
	if DEBUG >= glog.Level {
		glog.Log(DEBUG, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if DEBUG >= glog.Level {
		glog.Logf(DEBUG, format, args...)
	}
}

func Info(args ...interface{}) {
	if INFO >= glog.Level {
		glog.Log(INFO, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if INFO >= glog.Level {
		glog.Logf(INFO, format, args...)
	}
}

func Warn(args ...interface{}) {
	if WARN >= glog.Level {
		glog.Log(WARN, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if WARN >= glog.Level {
		glog.Logf(WARN, format, args...)
	}
}

func Error(args ...interface{}) {
	if ERROR >= glog.Level {
		glog.Log(ERROR, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if ERROR >= glog.Level {
		glog.Logf(ERROR, format, args...)
	}
}

func ErrorStack(format string, args ...interface{}) {
	if ERROR >= glog.Level {
		format = format + "\n%s"
		args = append(args, stack(false))
		glog.Logf(ERROR, format, args...)
	}
}

func Fatal(args ...interface{}) {
	if FATAL >= glog.Level {
		glog.Log(FATAL, args...)
		Close() // FATAL前关闭掉所有日志,避免损失丢失关键信息
		os.Exit(FATAL_EXIT_CODE)
	}
}

func Fatalf(format string, args ...interface{}) {
	if FATAL >= glog.Level {
		glog.Logf(FATAL, format, args...)
		Close() // FATAL前关闭掉所有日志,避免损失丢失关键信息
		os.Exit(FATAL_EXIT_CODE)
	}
}

func FatalStack(format string, args ...interface{}) {
	if FATAL >= glog.Level {
		format = format + "\n%s"
		args = append(args, stack(false))
		glog.Logf(FATAL, format, args...)
		Close() // FATAL前关闭掉所有日志,避免损失丢失关键信息
		os.Exit(FATAL_EXIT_CODE)
	}
}

func Flush() {
	if glog != nil {
		glog.Flush()
	}
	for _, v := range gmap {
		v.Flush()
	}
}

func Close() {
	if gcnf != nil {
		gcnf()
	}
	if glog != nil {
		glog.Close()
	}
	for _, v := range gmap {
		v.Close()
	}
}

func flush(c context.Context, flushPeriod time.Duration, g *Logger, m map[string]*Logger) {
	tick := time.Tick(flushPeriod)
	for {
		select {
		case <-c.Done(): // 结束
			return
		case <-tick:
			protectFlush(g, m)
		}
	}
}

func protectFlush(g *Logger, m map[string]*Logger) {
	defer func() {
		if perr := recover(); perr != nil {
			fmt.Fprintf(os.Stderr, "log flush panic: %v", perr)
		}
	}()

	if g != nil {
		g.Flush()
	}
	for _, v := range m {
		v.Flush()
	}
}
