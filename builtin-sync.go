package log

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type syncWriter struct {
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
}

func newSyncWriter(c *Config) (ret *syncWriter, err error) {
	var (
		rotateBytes int64
		rotateCycle Cycle
		file        *os.File
		size        int64
		year        int
		month       time.Month
		day         int
	)

	switch lpath := strings.ToLower(c.Path); lpath {
	case STDOUT:
		rotateBytes = 0
		rotateCycle = NEVER
		file = os.Stdout
	case STDERR:
		rotateBytes = 0
		rotateCycle = NEVER
		file = os.Stderr
	default:
		rotateBytes = c.RotateBytes
		rotateCycle = c.RotateCycle
		fi, _ := os.Stat(c.Path)
		if fi != nil {
			size = fi.Size()
			year, month, day = fi.ModTime().Date()
		} else {
			size = 0
			year, month, day = time.Now().Date()
		}

		file, err = os.OpenFile(c.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
	}

	ret = &syncWriter{
		path:            c.Path,
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
	}

	return
}
func (w *syncWriter) Log(level Level, msgs ...interface{}) {
	r := recordPool.Get().(*record) // 会在write方法中归还
	r.Buffer.Reset()
	printHeader(r, level, SKIP)
	fmt.Fprintln(r.Buffer, msgs...) // 不要用Fprint(), 会把相邻二个string拼接起来
	w.Write(r)
}

func (w *syncWriter) Logf(level Level, format string, args ...interface{}) {
	r := recordPool.Get().(*record) // 会在write方法中归还
	r.Buffer.Reset()
	printHeader(r, level, SKIP)
	fmt.Fprintf(r.Buffer, format, args...)
	r.Buffer.WriteByte('\n') // 没必要去比较, 大多数据情况是不会带换行符的
	w.Write(r)
}

func (w *syncWriter) Write(r *record) (err error) {
	size := int64(HEADER_BYTES + r.Len())
	w.mutex.Lock()
	if (w.rotateCycle == DAILY && (r.Year > w.year || r.Month > w.month || r.Day > w.day)) ||
		(w.rotateBytes > 0 && w.size+size > w.rotateBytes) ||
		(w.rotateCycle == MONTHLY && (r.Year > w.year || r.Month > w.month)) ||
		(w.rotateCycle == YEARLY && r.Year > w.year) {

		// 刷新关闭旧流
		w.writer.Flush()
		w.file.Close()

		// 重新命名旧文件
		rename(w.path, w.year, w.month, w.day)

		// 创建打开新流
		w.file, err = os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			goto __END
		}
		w.writer = bufio.NewWriterSize(w.file, w.bufioWriterSize)
		w.year, w.month, w.day = r.Year, r.Month, r.Day
		w.size = 0
	}
	w.writer.Write(r.Header)
	w.writer.Write(r.Buffer.Bytes())
	w.size += size
__END:
	w.mutex.Unlock()
	recordPool.Put(r)
	return
}

func (w *syncWriter) Flush() {
	w.mutex.Lock()
	w.writer.Flush()
	w.file.Sync()
	w.mutex.Unlock()
}

func (w *syncWriter) Close() {
	w.mutex.Lock()
	w.writer.Flush()
	// 不能关闭标准输出/标准错误
	if w.file != os.Stdout && w.file != os.Stderr {
		w.file.Close()
	}
	w.mutex.Unlock()
}
