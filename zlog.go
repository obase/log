package zlog

import (
	"golang.org/x/net/context"
	"math/rand"
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

	DEF_FLUSH_PERIOD        = 30 * time.Second //与glog相同
	DEF_RECORD_BUF_IDLE     = 256              //与glog相同
	DEF_RECORD_BUF_SIZE     = 1024             // 默认1K
	DEF_RECORD_HEADER_BYTES = 24               // 用于header最长的栈
	DEF_WRITER_BUF_SIZE     = 256 * 1024       //与glog相同, 256k

	SPACE = '\x20'
	COLON = ':'
	MINUS = '-'
	DOT   = '.'
	CRLF  = '\n'

	TRACEID = "#tid#"
)

var desc [6]string = [6]string{"[D]", "[I]", "[W]", "[E]", "[F]", "[O]"}
var hexs [16]byte = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func Debug(ctx context.Context, format string, args ...interface{}) {
	if Default.Level <= DEBUG {
		Default.printf(2, DEBUG, ctx, format, args...)
	}
}

func Info(ctx context.Context, format string, args ...interface{}) {
	if Default.Level <= INFO {
		Default.printf(2, INFO, ctx, format, args...)
	}
}

func Warn(ctx context.Context, format string, args ...interface{}) {
	if Default.Level <= WARN {
		Default.printf(2, WARN, ctx, format, args...)
	}
}

func Error(ctx context.Context, format string, args ...interface{}) {
	if Default.Level <= ERROR {
		Default.printf(2, ERROR, ctx, format, args...)
	}
}

func Fatal(ctx context.Context, format string, args ...interface{}) {
	if Default.Level <= FATAL {
		Default.printf(2, FATAL, ctx, format, args...)
	}
}

// call when program exit
func Flush() {
	if Default != nil {
		Default.Flush()
	}
	for _, v := range Others {
		v.Flush()
	}
}

type Option struct {
	Name          string `json:"name"` // 日志名字
	Level         Level  `json:"level"`
	Path          string `json:"path"`
	RotateBytes   int64  `json:"rotateBytes"`
	RotateCycle   Cycle  `json:"rotateCycle"`   //轮转周期,目前仅支持
	RecordBufIdle int    `json:"recordBufIdle"` // record buf的空闲数量
	RecordBufSize int    `json:"recordBufSize"` // record buf的初始大小
	WriterBufSize int    `json:"writerBufSize"` //Buffer写缓存大小
	Default       bool   `json:"default"`       //是否默认
}

// call this on init method
func flushDaemon(flushPeriod time.Duration) {
	for _ = range time.NewTicker(flushPeriod).C {
		Default.Flush()
		for _, v := range Others {
			v.Flush()
		}
	}
}

var (
	Default *logger
	Others  map[string]*logger = make(map[string]*logger)
)

func GetLog(name string) (l *logger) {
	l, ok := Others[name]
	if !ok {
		l = Default
	}
	return
}

func Init(flushPeriod time.Duration, opts ...*Option) (err error) {

	if flushPeriod <= 0 {
		flushPeriod = DEF_FLUSH_PERIOD
	}

	// 如果错误,需要关闭相应的文件句柄
	defer func() {
		if err != nil {
			if Default != nil {
				Default.File.Close()
			}
			for _, v := range Others {
				v.File.Close()
			}
		}
	}()

	var lgr *logger
	// 初始化全局变量
	for _, opt := range opts {
		lgr, err = newLogger(opt)
		if opt.Default {
			Default = lgr
		} else {
			Others[opt.Name] = lgr
		}
	}
	// 启动定期刷新么台
	go flushDaemon(flushPeriod)
	return
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

// 使用时间做seed确保每次的序列不同
var zrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandTraceId() string {

	bs := make([]byte, 20)

	rd := zrand.Uint64()
	for i := 19; i >= 4; i-- {
		bs[i] = hexs[rd&0x0f]
		rd >>= 4
	}

	ts := time.Now().UnixNano()
	for i := 3; i >= 0; i-- {
		bs[i] = hexs[ts&0x0f]
		ts >>= 4
	}

	return string(bs)
}
