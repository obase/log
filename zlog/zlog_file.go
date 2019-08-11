package zlog

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

type record struct {
	bytes.Buffer
	Header []byte
}

var recordSyncPool = sync.Pool{
	New: func() interface{} {
		return &record{
			Header: make([]byte, RECORD_HEADER_BYTES),
		}
	},
}

type Logger struct {
	*Config
	*os.File
	*bufio.Writer
	sync.Mutex
	Year  int
	Month time.Month
	Day   int
	Hour  int
	Size  int64
}

func newLogger(c *Config) (*Logger, error) {
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
		return nil, err
	}

	now := time.Now()
	nyear, nmonth, nday := now.Date()
	nhour := now.Hour()

	if info == nil {
		dir := filepath.Dir(c.Path)
		if info, err = os.Stat(dir); err != nil {
			return nil, err
		}
		if info == nil {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return nil, err
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
				return nil, err
			}
			year, month, day, hour, size = nyear, nmonth, nday, nhour, 0
		}
	}
	if file, err = openFile(c.Path); err != nil {
		return nil, err
	}
	return &Logger{
		Config: c,
		File:   file,
		Writer: bufio.NewWriterSize(file, c.BufioWriterSize),
		Year:   year,
		Month:  month,
		Day:    day,
		Hour:   hour,
		Size:   size,
	}, nil
}

func openFile(path string) (*os.File, error) {
	if path == STDOUT {
		return os.Stdout, nil
	} else if path == STDERR {
		return os.Stderr, nil
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
}
func rename(path string, year int, month time.Month, day int, hour int) error {

	buf := recordSyncPool.Get().(*record)
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
	buf.Header[3], buf.Header[2], buf.Header[1], buf.Header[0] = hexs[month%10], hexs[month/10%10], hexs[month/100%10], hexs[month/1000%10]

	buf.Write(buf.Header[:13])
	npath := buf.String()
	recordSyncPool.Put(buf)

	nsize := len(npath)
	for i := 1; i < 256; i++ {
		if info, _ := os.Stat(npath); info != nil {
			npath = npath[:nsize] + "." + strconv.Itoa(i)
		} else {
			return os.Rename(path, npath)
		}
	}
	return errors.New("too much file with same prefix: " + npath)
}

func (l *Logger) printf(level Level, format string, args []interface{}) {

	_, file, line, ok := runtime.Caller(2) // 调用链深度
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

	buf := recordSyncPool.Get().(*record)
	// yyyy-MM-dd HH:mm:ss.SSS共23个字符
	buf.Header[22], buf.Header[21], buf.Header[20] = hexs[millis%10], hexs[millis/10%10], hexs[millis/100%10]
	buf.Header[19] = DOT
	buf.Header[18], buf.Header[17] = hexs[second%10], hexs[second/10%10]
	buf.Header[16] = COLON
	buf.Header[15], buf.Header[14] = hexs[minute%10], hexs[minute/10%10]
	buf.Header[13] = COLON
	buf.Header[12], buf.Header[11] = hexs[hour%10], hexs[hour/10%10]
	buf.Header[10] = SPACE
	buf.Header[9], buf.Header[8] = hexs[day%10], hexs[day/10%10]
	buf.Header[7] = MINUS
	buf.Header[6], buf.Header[5] = hexs[month%10], hexs[month/10%10]
	buf.Header[4] = MINUS
	buf.Header[3], buf.Header[2], buf.Header[1], buf.Header[0] = hexs[year%10], hexs[year/10%10], hexs[year/100%10], hexs[year/1000%10]
	buf.Write(buf.Header[:23])
	buf.WriteByte(SPACE)
	buf.WriteString(desc[level])
	buf.WriteByte(SPACE)
	buf.WriteString(file)
	buf.WriteByte(COLON)

	idx := 10 // line最多10个字符
	for line > 0 {
		idx--
		buf.Header[idx] = hexs[line%10];
		line /= 10
	}
	buf.Write(buf.Header[idx:10])
	buf.WriteByte(SPACE)

	fmt.Fprintf(buf, format, args...)
	if format[len(format)-1] != CRLF {
		buf.WriteByte(CRLF)
	}
	size := int64(buf.Len())

	l.Mutex.Lock()
	rotated := false
	if l.Config.RotateBytes > 0 && l.Config.RotateBytes < l.Size+size {
		rotated = true
	} else {
		switch l.Config.RotateCycle {
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
	if rotated {
		l.Writer.Flush() //刷新缓存
		l.File.Close()   // 关闭句柄
		rename(l.Config.Path, l.Year, l.Month, l.Day, l.Hour)
		l.File, _ = os.OpenFile(l.Config.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		l.Writer = bufio.NewWriterSize(l.File, l.Config.BufioWriterSize)
		l.Year, l.Month, l.Day, l.Hour, l.Size = year, month, day, hour, 0
	}
	l.Writer.Write(buf.Bytes())
	l.Mutex.Unlock()

	recordSyncPool.Put(buf)
}

func (l *Logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.Config.Level <= DEBUG {
		l.printf(DEBUG, format, args)
	}
}

func (l *Logger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.Config.Level <= INFO {
		l.printf(INFO, format, args)
	}
}

func (l *Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.Config.Level <= WARN {
		l.printf(WARN, format, args)
	}
}

func (l *Logger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.Config.Level <= ERROR {
		l.printf(ERROR, format, args)
	}
}

func (l *Logger) Fatal(ctx context.Context, format string, args ...interface{}) {
	if l.Config.Level <= FATAL {
		l.printf(FATAL, format, args)
	}
}

func (l *Logger) Flush() {
	l.Writer.Flush()
	l.File.Sync()
}
