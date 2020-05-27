package log

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type asyncWriter struct {
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
	lctx            context.Context // 异步读写
	lcnf            context.CancelFunc
	lchn            chan *record
	closeDelay      time.Duration
}

func newAsyncWriter(c *Config) (ret *asyncWriter, err error) {
	var (
		rotateBytes int64
		rotateCycle Cycle
		file        *os.File
		size        int64
		year        int
		month       time.Month
		day         int
		lctx        context.Context
		lcnf        context.CancelFunc
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

	lctx, lcnf = context.WithCancel(context.Background())

	ret = &asyncWriter{
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
		lctx:            lctx,
		lcnf:            lcnf,
		lchn:            make(chan *record, c.AsyncWriteLimit),
		closeDelay:      c.AsyncCloseDelay,
	}

	go ret.launchAsyncProcess()

	return
}
func (w *asyncWriter) Log(level Level, msgs ...interface{}) {
	r := recordPool.Get().(*record) // 会在write方法中归还
	r.Buffer.Reset()
	printHeader(r, level, SKIP)
	fmt.Fprintln(r.Buffer, msgs...) // 不要用Fprint(), 会把相邻二个string拼接起来
	w.lchn <- r
}

func (w *asyncWriter) Logf(level Level, format string, args ...interface{}) {
	r := recordPool.Get().(*record) // 会在write方法中归还
	r.Buffer.Reset()
	printHeader(r, level, SKIP)
	fmt.Fprintf(r.Buffer, format, args...)
	r.Buffer.WriteByte('\n') // 没必要去比较, 大多数据情况是不会带换行符的
	w.lchn <- r
}

func (w *asyncWriter) launchAsyncProcess() {
	for { // 启动后台异步进程, 非正常退出则自动重入
		exit := func() bool {
			defer func() {
				if perr := recover(); perr != nil {
					fmt.Fprintf(os.Stderr, "async process panic: %v", perr)
				}
			}()
			for {
				select {
				case r := <-w.lchn:
					w.Write(r)
				case <-w.lctx.Done():
					return true
				}
			}
		}()
		if exit {
			return
		}
	}
}

func (w *asyncWriter) Write(r *record) (err error) {
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

func (w *asyncWriter) Flush() {
	w.mutex.Lock()
	w.writer.Flush()
	w.file.Sync()
	w.mutex.Unlock()
}

func (w *asyncWriter) Close() {
	// 如果是异步稍做等待
	if w.file != nil {
		time.Sleep(w.closeDelay)
	}
	w.mutex.Lock()
	w.lcnf() // 关闭异步上下文
	w.writer.Flush()
	// 不能关闭标准输出/标准错误
	if w.file != os.Stdout && w.file != os.Stderr {
		w.file.Close()
	}
	w.mutex.Unlock()
}
