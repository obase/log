package log

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
	OFF  //不输出任何日志
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

type Buffer struct {
	bytes.Buffer
	Header []byte
}

var bufferSyncPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{
			Header: make([]byte, RECORD_HEADER_BYTES),
		}
	},
}

type Logger struct {
	sync.Mutex
	Year    int
	Month   time.Month
	Day     int
	Hour    int
	Size    int64
	Rotated bool // 需要轮转
	*Config
	*os.File
	*bufio.Writer
}

const SKIP int = 2; // 统一跳过函数栈层次

func (l *Logger) log(level Level, format string, args []interface{}) {
	_, file, line, ok := runtime.Caller(SKIP) // 调用链深度
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

	r := bufferSyncPool.Get().(*Buffer)
	defer bufferSyncPool.Put(r)
	// 先重置
	r.Reset()
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

	fmt.Fprintf(r, format, args...)
	if format[len(format)-1] != CRLF {
		r.WriteByte(CRLF)
	}

	l.Mutex.Lock()
	{
		size := int64(r.Len())
		rotated := false
		if l.RotateBytes > 0 && l.RotateBytes < l.Size+size {
			rotated = true
		} else {
			switch l.RotateCycle {
			case NONE:
				rotated = false
			case YEARLY:
				rotated = (l.Year != year)
			case MONTHLY:
				rotated = (l.Year != year) || (l.Month != month)
			case DAILY:
				rotated = (l.Year != year) || (l.Month != month) || (l.Day != day)
			case HOURLY:
				rotated = (l.Year != year) || (l.Month != month) || (l.Day != day) || (l.Hour != hour)
			}
		}
		if rotated && l.Rotated {
			l.Writer.Flush() //刷新缓存
			l.File.Close()   // 关闭句柄
			rename(l.Path, l.Year, l.Month, l.Day, l.Hour)
			l.File, _ = os.OpenFile(l.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			l.Writer = bufio.NewWriterSize(l.File, l.BufioWriterSize)
			l.Year, l.Month, l.Day, l.Hour, l.Size = year, month, day, hour, 0
		}
		l.Writer.Write(r.Bytes())
		l.Size += size
	}
	l.Mutex.Unlock()
}

func (l *Logger) Flush() {
	l.Mutex.Lock()
	{
		l.Writer.Flush()
		l.File.Sync()
	}
	l.Mutex.Unlock()
}

func (l *Logger) Close() {
	l.Mutex.Lock()
	{
		l.Writer.Flush()
		l.File.Close()
	}
	l.Mutex.Unlock()
}

func (l *Logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= DEBUG {
		l.log(DEBUG, format, args)
	}
}

func (l *Logger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= INFO {
		l.log(INFO, format, args)
	}
}

func (l *Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= WARN {
		l.log(WARN, format, args)
	}
}

func (l *Logger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= ERROR {
		l.log(ERROR, format, args)
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
		l.log(ERROR, "%v\n%s", []interface{}{err, stacks(all)})
	}
}

func (l *Logger) Fatal(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= FATAL {
		l.log(FATAL, format, args)
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

		var logger *Logger
		switch c.Path {
		case STDOUT: // 不需要轮转
			logger = &Logger{
				Config:  c,
				Rotated: false,
				File:    os.Stdout,
				Writer:  bufio.NewWriterSize(os.Stdout, c.BufioWriterSize),
			}
		case STDERR: // 不需要轮转
			logger = &Logger{
				Config:  c,
				Rotated: false,
				File:    os.Stderr,
				Writer:  bufio.NewWriterSize(os.Stderr, c.BufioWriterSize),
			}
		default: // 需要轮转
			var (
				file  *os.File
				info  os.FileInfo
				err   error
				year  int
				month time.Month
				day   int
				hour  int
				size  int64
			)

			if info, err = os.Stat(c.Path); err != nil {
				if !os.IsNotExist(err) && !os.IsExist(err) {
					return err
				}
			}

			now := time.Now()
			nyear, nmonth, nday := now.Date()
			nhour := now.Hour()

			if info == nil {
				dir := filepath.Dir(c.Path)
				if info, err = os.Stat(dir); err != nil {
					if !os.IsNotExist(err) && !os.IsExist(err) {
						return err
					}
				}
				if info == nil {
					if err = os.MkdirAll(dir, os.ModePerm); err != nil {
						return err
					}
				}
				year, month, day, hour, size = nyear, nmonth, nday, nhour, 0

			} else {

				size = info.Size()
				mtime := info.ModTime()
				year, month, day = mtime.Date()
				hour = mtime.Hour()
				size = info.Size()

				rotated := false
				if c.RotateBytes > 0 && c.RotateBytes < size {
					rotated = true
				} else {
					switch c.RotateCycle {
					case NONE:
						rotated = false
					case YEARLY:
						rotated = (nyear != year)
					case MONTHLY:
						rotated = (nyear != year) || (nmonth != month)
					case DAILY:
						rotated = (nyear != year) || (nmonth != month) || (nday != day)
					case HOURLY:
						rotated = (nyear != year) || (nmonth != month) || (nday != day) || (nhour != hour)
					}
				}
				if rotated {
					if err = rename(c.Path, year, month, day, hour); err != nil {
						return err
					}
					year, month, day, hour, size = nyear, nmonth, nday, nhour, 0
				}
			}
			if file, err = os.OpenFile(c.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
				return err
			}
			logger = &Logger{
				Config:  c,
				Rotated: true,
				File:    file,
				Writer:  bufio.NewWriterSize(file, c.BufioWriterSize),
				Year:    year,
				Month:   month,
				Day:     day,
				Hour:    hour,
				Size:    size,
			}
		}

		if c.Name != "" {
			for _, k := range strings.Split(c.Name, ",") {
				_loggers[k] = logger
			}
		}
		if c.Default {
			_default = logger
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

func rename(path string, year int, month time.Month, day int, hour int) error {

	buf := bufferSyncPool.Get().(*Buffer)
	defer bufferSyncPool.Put(buf)

	buf.Reset()
	buf.WriteString(path)
	buf.WriteByte(DOT)

	// yyyy-MM-dd.HH共13个字符
	buf.Header[12], buf.Header[11] = hexs[hour%10], hexs[hour/10%10]
	buf.Header[10] = DOT
	buf.Header[9], buf.Header[8] = hexs[day%10], hexs[day/10%10]
	buf.Header[7] = MINUS
	buf.Header[6], buf.Header[5] = hexs[month%10], hexs[month/10%10]
	buf.Header[4] = MINUS
	buf.Header[3], buf.Header[2], buf.Header[1], buf.Header[0] = hexs[year%10], hexs[year/10%10], hexs[year/100%10], hexs[year/1000%10]

	buf.Write(buf.Header[:13])
	npath := buf.String()

	nsize := len(npath)
	for i := 1; i < 65535; i++ {
		if info, _ := os.Stat(npath); info != nil {
			npath = npath[:nsize] + "." + strconv.Itoa(i)
		} else {
			return os.Rename(path, npath)
		}
	}
	return errors.New("too much file with same prefix: " + npath)
}
