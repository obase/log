package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
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

	DEF_FLUSH_PERIOD = 5 * time.Second //与glog相同

	DEF_BUFIO_WRITER_SIZE = 256 * 1024 //与glog相同, 256k

	RECORD_HEADER_BYTES = 24 // 用于header最长的栈
	SPACE               = '\x20'
	COLON               = ':'
	MINUS               = '-'
	DOT                 = '.'
	CRLF                = '\n'
)

var desc [6]string = [6]string{"[D]", "[I]", "[W]", "[E]", "[F]", "[O]"}
var hexs [16]byte = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

type Config struct {
	Name            string // 日志名字
	Level           Level
	Path            string
	RotateBytes     int64
	RotateCycle     Cycle //轮转周期,目前仅支持
	BufioWriterSize int   //Buffer写缓存大小
	Default         bool  //是否默认
}

type Record struct {
	bytes.Buffer
	Header []byte
	Year   int
	Month  time.Month
	Day    int
	Hour   int
}

var recordSyncPool = sync.Pool{
	New: func() interface{} {
		return &Record{
			Header: make([]byte, RECORD_HEADER_BYTES),
		}
	},
}

type Writer interface {
	Write(r *Record)
	Flush()
	Close()
}

type Logger struct {
	Writer
	Level Level
}

const SKIP int = 2 // 统一跳过函数栈层次

func (l *Logger) Log(skip int, level Level, format string, args []interface{}) {
	_, file, line, ok := runtime.Caller(skip) // 调用链深度
	if !ok {
		file = "???"
		line = 1
	} else if pos := strings.LastIndexByte(file, '/'); pos >= 0 {
		file = file[pos+1:]
	}

	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	millis := now.Nanosecond() / 1000000

	r := recordSyncPool.Get().(*Record)
	defer recordSyncPool.Put(r)

	r.Reset() // 先重置
	// 标记时间点
	r.Year, r.Month, r.Day, r.Hour = year, month, day, hour

	// yyyy-MM-dd HH:mm:ss.SSS共23个字符
	r.Header[22], r.Header[21], r.Header[20] = hexs[millis%10], hexs[millis/10%10], hexs[millis/100%10]
	r.Header[19] = DOT
	r.Header[18], r.Header[17] = hexs[second%10], hexs[second/10%10]
	r.Header[16] = COLON
	r.Header[15], r.Header[14] = hexs[minute%10], hexs[minute/10%10]
	r.Header[13] = COLON
	r.Header[12], r.Header[11] = hexs[hour%10], hexs[hour/10%10]
	r.Header[10] = SPACE
	r.Header[9], r.Header[8] = hexs[day%10], hexs[day/10%10]
	r.Header[7] = MINUS
	r.Header[6], r.Header[5] = hexs[month%10], hexs[month/10%10]
	r.Header[4] = MINUS
	r.Header[3], r.Header[2], r.Header[1], r.Header[0] = hexs[year%10], hexs[year/10%10], hexs[year/100%10], hexs[year/1000%10]
	r.Write(r.Header[:23])
	r.WriteByte(SPACE)
	r.WriteString(desc[level])
	r.WriteByte(SPACE)
	r.WriteString(file)
	r.WriteByte(COLON)

	idx := 10 // line最多10个字符
	for line > 0 {
		idx--
		r.Header[idx] = hexs[line%10]
		line /= 10
	}
	r.Write(r.Header[idx:10])
	r.WriteByte(SPACE)
	r.WriteByte(MINUS)
	r.WriteByte(SPACE)

	// 优化无需格式的情况
	if len(args) > 0 {
		fmt.Fprintf(r, format, args...)
	} else {
		r.WriteString(format)
	}

	if format[len(format)-1] != CRLF {
		r.WriteByte(CRLF)
	}

	l.Writer.Write(r) // 是否需要rotate由writer自己实现
}

func (l *Logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= DEBUG {
		l.Log(SKIP, DEBUG, format, args)
	}
}

func (l *Logger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= INFO {
		l.Log(SKIP, INFO, format, args)
	}
}

func (l *Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= WARN {
		l.Log(SKIP, WARN, format, args)
	}
}

func (l *Logger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= ERROR {
		l.Log(SKIP, ERROR, format, args)
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

func (l *Logger) ErrorStack(ctx context.Context, err interface{}, all bool) {
	if l.Level <= ERROR {
		l.Log(SKIP, ERROR, "%v\n%s", []interface{}{err, stacks(all)})
	}
}

func (l *Logger) Fatal(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= FATAL {
		l.Log(SKIP, FATAL, format, args)
		os.Exit(1)
	}
}

var (
	_default *Logger
	_loggers map[string]*Logger = make(map[string]*Logger)
)

func GetLog(key string) *Logger {
	return _loggers[key]
}

func Setup(flushPeriod time.Duration, configs ...*Config) (err error) {

	// 初始化全局变量
	for _, c := range configs {

		c = mergeConfig(c)

		var writer Writer
		if c.Path == STDOUT {
			writer = NewConsoleWriter(os.Stdout)
		} else if c.Path == STDERR {
			writer = NewConsoleWriter(os.Stderr)
		} else {
			if writer, err = NewTextfileWriter(c); err != nil {
				if _default != nil {
					_default.Writer.Close()
				}
				for _, v := range _loggers {
					v.Writer.Close()
				}
				return err
			}
		}
		var logger = &Logger{
			Writer: writer,
			Level:  c.Level,
		}
		if c.Name != "" {
			for _, k := range strings.Split(c.Name, ",") {
				_loggers[k] = logger
			}
		}
		if c.Default {
			_default = logger
			_loggers[""] = logger // 更好集成到其他平台
		}
	}

	// 启动定期刷新么台
	go flushDaemon(flushPeriod)

	return nil
}

func mergeConfig(c *Config) *Config {
	if c == nil {
		c = new(Config)
	}
	if c.Path == "" {
		c.Path = STDOUT
	}
	if c.BufioWriterSize <= 0 {
		c.BufioWriterSize = DEF_BUFIO_WRITER_SIZE
	}
	return c
}

const DefaultFlushInterval = 30 * time.Second

func flushDaemon(flushInterval time.Duration) {
	if flushInterval <= 0 {
		flushInterval = DefaultFlushInterval
	}
	for _ = range time.Tick(flushInterval) {
		_default.Flush()
		for _, v := range _loggers {
			v.Flush()
		}
	}
}
