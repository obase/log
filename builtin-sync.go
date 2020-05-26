package log

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"
)

type SyncWriter struct {
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

func newSyncWriter(c *Config) (ret *SyncWriter, err error) {
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

	ret = &SyncWriter{
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
	}

	return
}
func (w *SyncWriter) Log(level Level, msgs ...interface{}) {
	return
}

func (w *SyncWriter) Logf(level Level, format string, args ...interface{}) {
	return
}

func (w *SyncWriter) Write(r *Record) (err error) {
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
	return
}

func (w *SyncWriter) Flush() {
	w.mutex.Lock()
	w.writer.Flush()
	w.file.Sync()
	w.mutex.Unlock()
}

func (w *SyncWriter) Close() {
	w.mutex.Lock()
	w.writer.Flush()
	w.file.Close()
	w.mutex.Unlock()
}
