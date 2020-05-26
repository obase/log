package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Cycle uint

const (
	NEVER Cycle = iota
	DAILY
	MONTHLY
	YEARLY
)

const (
	STDOUT = "stdout"
	STDERR = "stderr"
)

type Config struct {
	Level           Level
	Path            string
	RotateBytes     int64
	RotateCycle     Cycle //轮转周期,目前仅支持
	BufioWriterSize int   //Buffer写缓存大小
	Async           bool  // 是否使用异步
	AsyncLimit      int   // 异步管道容量,默认512
}

func MergeConfig(c *Config) *Config {
	if c == nil {
		c = new(Config)
	}
	if c.Path == "" {
		c.Path = STDOUT
	}
	if c.BufioWriterSize == 0 {
		c.BufioWriterSize = 256 * 1024 // 默认256K
	}
	if c.AsyncLimit == 0 {
		c.AsyncLimit = 512
	}
	return c
}

const (
	DOT   = '.'
	MINUS = '-'
)

const (
	HEADER_BYTES = 28   // 用于保存header部分"2020-05-26 16:40:11.370 [I] "
	BUFFER_BYTES = 1024 // 初始buffer的大小, 默认1K
)

type Record struct {
	*bytes.Buffer
	Header []byte
	Year   int
	Month  time.Month
	Day    int
}

var buffPool = sync.Pool{
	New: func() interface{} {
		return &Record{
			Buffer: bytes.NewBuffer(make([]byte, 0, BUFFER_BYTES)),
			Header: make([]byte, HEADER_BYTES),
		}
	},
}

func BorrowRecord() *Record {
	return buffPool.Get().(*Record)
}

func ReturnRecord(buf *Record) {
	buffPool.Put(buf)
}

func newBuiltinLogger(c *Config) (ret *Logger, err error) {

	c = MergeConfig(c)

	if c.Async {
		writer, err := newAsyncWriter(c)
		if err != nil {
			return
		}
		ret = &Logger{
			Log:   writer.Log,
			Logf:  writer.Logf,
			Flush: writer.Flush,
			Close: writer.Close,
		}
	} else {
		writer, err := newSyncWriter(c)
		if err != nil {
			return
		}
		ret = &Logger{
			Log:   writer.Log,
			Logf:  writer.Logf,
			Flush: writer.Flush,
			Close: writer.Close,
		}
	}
	return
}

var desc [5]byte = [5]byte{'D', 'I', 'W', 'E', 'F'}
var hexs [16]byte = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func rename(path string, year int, month time.Month, day int) {

	buf := BorrowRecord()
	buf.Buffer.Reset()
	buf.Buffer.WriteString(path)
	buf.Buffer.WriteByte(DOT)

	// yyyy-MM-dd.HH共13个字符

	buf.Header[9] = hexs[day%10]
	buf.Header[8] = hexs[day/10]
	buf.Header[7] = MINUS
	buf.Header[6] = hexs[month%10]
	buf.Header[5] = hexs[month/10]
	buf.Header[4] = MINUS
	buf.Header[3] = hexs[year%10]
	year /= 10
	buf.Header[2] = hexs[year%10]
	year /= 10
	buf.Header[1] = hexs[year%10]
	year /= 10
	buf.Header[0] = hexs[year]

	buf.Buffer.Write(buf.Header[:10])
	buf.Buffer.WriteByte(DOT)
	nsize := buf.Buffer.Len()
	buf.Buffer.WriteString(strconv.FormatInt(time.Now().UnixNano(), 36))
	npath := buf.String()
	ReturnRecord(buf)

	for {
		if info, err := os.Stat(path); info != nil || os.IsExist(err) {
			npath = npath[:nsize] + strconv.FormatInt(time.Now().UnixNano(), 36) // 重新拼过时间戳
		} else {
			if err := os.Rename(path, npath); err != nil {
				fmt.Fprintf(os.Stderr, "rename log file error: %v, %v -> %v\n", err, path, npath)
			}
			return
		}
	}

}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stack(all bool) []byte {
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
