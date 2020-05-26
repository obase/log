package log

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
	"time"
)

type AsyncWriter struct {
	path            string
	bufioWriterSize int
	rotateBytes     int64
	rotateCycle     Cycle
	file            *os.File
	writer          *bufio.Writer
	mutex           *sync.Mutex
	size            int64
	year            int
	month           time.Month
	day             int
	// 异步读写
	contex       context.Context
	cancel       context.CancelFunc
	asynChanSize int // 异步通道大小
}

func newAsyncWriter(c *Config) (ret *AsyncWriter, err error) {
	var (
		path        string
		file        *os.File
		rotateBytes int64
		rotateCycle Cycle
		size        int64
		year        int
		month       time.Month
		day         int
	)

	switch lpath := strings.ToLower(c.Path); lpath {
	case STDOUT:
		path = lpath
		file = os.Stdout
	case STDERR:
		path = lpath
		file = os.Stderr
	default:
		path = c.Path
		rotateBytes = c.RotateBytes
		rotateCycle = c.RotateCycle
		file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		fi, _ := os.Stat(path)
		if fi != nil {
			size = fi.Size()
			mtime := fi.ModTime()
			year, month, day = mtime.Date()
		}
	}

	ctx, cfn := context.WithCancel(context.Background())

	ret = &AsyncWriter{
		path:            path,
		bufioWriterSize: c.BufioWriterSize,
		rotateCycle:     rotateCycle,
		rotateBytes:     rotateBytes,
		file:            file,
		writer:          bufio.NewWriterSize(file, c.BufioWriterSize),
		mutex:           new(sync.Mutex),
		size:            size,
		year:            year,
		month:           month,
		day:             day,
		contex:          ctx,
		cancel:          cfn,
	}

	return
}
func (w *AsyncWriter) Log(level Level, msgs ...interface{}) {
	return
}

func (w *AsyncWriter) Logf(level Level, format string, args ...interface{}) {
	return
}

func (w *AsyncWriter) write(r *Record) {

}

func (w *AsyncWriter) Flush() {
	return
}

func (w *AsyncWriter) Close() {
	return
}
