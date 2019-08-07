package log

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type Level uint8
type Cycle uint8

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
	OFF //不输出任何日志
)

const (
	NONE Cycle = iota
	HOURLY
	DAILY
	MONTHLY
	YEARLY
)

const (
	STDOUT = "stdout"
	STDERR = "stderr"

	DEF_FLUSH_PERIOD        = 5 * time.Second //与glog相同
	DEF_RECORD_BUF_IDLE     = 256             //与glog相同
	DEF_RECORD_BUF_SIZE     = 1024            // 默认1K
	DEF_RECORD_HEADER_BYTES = 24              // 用于header最长的栈
	DEF_WRITER_BUF_SIZE     = 256 * 1024      //与glog相同, 256k

	SPACE = '\x20'
	COLON = ':'
	MINUS = '-'
	DOT   = '.'
	CRLF  = '\n'

	TRACEID = "#tid#"
)

var desc [6]string = [6]string{"[D]", "[I]", "[W]", "[E]", "[F]", "[O]"}
var hexs [16]byte = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

type Config struct {
	Name          string `json:"name" bson:"name" yaml:"name"` // 日志名字
	Level         Level  `json:"level" bson:"level" yaml:"level"`
	Path          string `json:"path" bson:"path" yaml:"path"`
	RotateBytes   int64  `json:"rotateBytes" bson:"rotateBytes" yaml:"rotateBytes"`
	RotateCycle   Cycle  `json:"rotateCycle" bson:"rotateCycle" yaml:"rotateCycle"`       //轮转周期,目前仅支持
	RecordBufIdle int    `json:"recordBufIdle" bson:"recordBufIdle" yaml:"recordBufIdle"` // record buf的空闲数量
	RecordBufSize int    `json:"recordBufSize" bson:"recordBufSize" yaml:"recordBufSize"` // record buf的初始大小
	WriterBufSize int    `json:"writerBufSize" bson:"writerBufSize" yaml:"writerBufSize"` //Buffer写缓存大小
	Default       bool   `json:"default" bson:"default" yaml:"default"`                   //是否默认
}

// call this on init method
func flushDaemon(flushPeriod time.Duration) {
	defer func() {
		if perr := recover(); perr != nil {
			fmt.Fprintf(os.Stderr, "log flushDaemon error: %v\n", perr)
		}
	}()
	for _ = range time.NewTicker(flushPeriod).C {
		_default.Flush()
		for _, v := range _loggers {
			v.Flush()
		}
	}
}

var (
	_default *logger
	_loggers map[string]*logger = make(map[string]*logger)
)

func GetLog(name string) (l *logger) {
	if rt, ok := _loggers[name]; ok {
		return rt
	}
	return nil
}

func Setup(flushPeriod time.Duration, opts ...*Config) (err error) {

	// 如果错误,需要关闭相应的文件句柄. 否则重置全局的函数变量
	defer func() {
		if err != nil {
			if _default != nil {
				_default.File.Close()
			}
			for _, v := range _loggers {
				v.File.Close()
			}
		} else if _default != nil {
			// 设置def里面函数变量
			Fatal = _default.Fatal
			Error = _default.Error
			Warn = _default.Warn
			Info = _default.Info
			Debug = _default.Debug
			// 兼容旧版本
			Fatalf = _default.Fatal
			Errorf = _default.Error
			Warnf = _default.Warn
			Infof = _default.Info
			Inforf = _default.Info
			Debugf = _default.Debug
		}
	}()

	if flushPeriod <= 0 {
		flushPeriod = DEF_FLUSH_PERIOD
	}

	var lgr *logger
	// 初始化全局变量
	for _, opt := range opts {
		if lgr, err = newLogger(mergeConfig(opt)); err != nil {
			return
		}
		if opt.Default {
			_default = lgr
		} else {
			for _, k := range strings.Split(opt.Name, ",") {
				_loggers[k] = lgr
			}
		}
	}

	// 启动定期刷新么台
	go flushDaemon(flushPeriod)
	return
}

func mergeConfig(c *Config) *Config {
	if c.Path == "" {
		c.Path = STDOUT
	}
	if c.RotateCycle == NONE {
		c.RotateCycle = DAILY
	}

	return c
}

func (l *logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= DEBUG {
		l.printf(2, DEBUG, ctx, format, args...)
	}
}

func (l *logger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= INFO {
		l.printf(2, INFO, ctx, format, args...)
	}
}

func (l *logger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= WARN {
		l.printf(2, WARN, ctx, format, args...)
	}
}

func (l *logger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= ERROR {
		l.printf(2, ERROR, ctx, format, args...)
	}
}

func (l *logger) Fatal(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= FATAL {
		l.printf(2, FATAL, ctx, format, args...)
	}
}
