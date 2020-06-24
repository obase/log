package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
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
}

func mergeConfig(c *Config) *Config {
	if c == nil {
		c = new(Config)
	}
	if c.Path == "" {
		c.Path = STDOUT
	}
	if c.BufioWriterSize == 0 {
		c.BufioWriterSize = 256 * 1024 // 默认256K
	}
	return c
}

const (
	DOT   = '.'
	MINUS = '-'
	SPACE = ' '
	COLON = ':'
)

const (
	HEADER_BYTES    = 28   // 用于保存header部分"2020-05-26 16:40:11.370 [I] "
	BUFFER_BYTES    = 1024 // 初始buffer的大小, 默认1K
	SKIP            = 3
	FATAL_EXIT_CODE = 7
)

var desc [5]byte = [5]byte{'D', 'I', 'W', 'E', 'F'}
var hexs [16]byte = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

type record struct {
	*bytes.Buffer
	Header []byte
	Year   int
	Month  time.Month
	Day    int
}

func newRecord() *record {
	return &record{
		Buffer: bytes.NewBuffer(make([]byte, 0, BUFFER_BYTES)),
		Header: make([]byte, HEADER_BYTES),
	}
}

var recordPool = sync.Pool{
	New: func() interface{} {
		return newRecord()
	},
}

func printHeader(r *record, level Level, skip int) *record {

	// 先写入文件行号(重用Header)
	_, file, line, ok := runtime.Caller(skip) // 调用链深度
	if !ok {
		file = "???"
		line = 1
	} else if pos := strings.LastIndexByte(file, '/'); pos >= 0 {
		file = file[pos+1:]
	}

	r.Buffer.WriteString(file)
	r.Buffer.WriteByte(COLON)

	mark := 0
	for line > 0 {
		mark++
		r.Header[mark] = hexs[line%10]
		line /= 10
	}
	for mark > 0 {
		r.Buffer.WriteByte(r.Header[mark])
		mark--
	}
	r.Buffer.WriteByte(SPACE)
	r.Buffer.WriteByte(MINUS)
	r.Buffer.WriteByte(SPACE)

	now := time.Now()
	yr, mn, dt := now.Date()
	hr, mi, sc := now.Clock()
	ms := now.Nanosecond() / 1000000

	// 再赋值年/月/日
	r.Year, r.Month, r.Day = yr, mn, dt
	// 最后渲染头部28个字符
	r.Header[27] = SPACE
	r.Header[26] = ']'
	r.Header[25] = desc[level]
	r.Header[24] = '['
	r.Header[23] = SPACE
	r.Header[22] = hexs[ms%10]
	ms /= 10
	r.Header[21] = hexs[ms%10]
	ms /= 10
	r.Header[20] = hexs[ms%10]
	r.Header[19] = DOT
	r.Header[18] = hexs[sc%10]
	sc /= 10
	r.Header[17] = hexs[sc%10]
	r.Header[16] = COLON
	r.Header[15] = hexs[mi%10]
	mi /= 10
	r.Header[14] = hexs[mi%10]
	r.Header[13] = COLON
	r.Header[12] = hexs[hr%10]
	hr /= 10
	r.Header[11] = hexs[hr%10]
	r.Header[10] = SPACE
	r.Header[9] = hexs[dt%10]
	dt /= 10
	r.Header[8] = hexs[dt%10]
	r.Header[7] = MINUS
	r.Header[6] = hexs[mn%10]
	mn /= 10
	r.Header[5] = hexs[mn%10]
	r.Header[4] = MINUS
	r.Header[3] = hexs[yr%10]
	yr /= 10
	r.Header[2] = hexs[yr%10]
	yr /= 10
	r.Header[1] = hexs[yr%10]
	yr /= 10
	r.Header[0] = hexs[yr%10]

	return r
}

func NewBuiltinLogger(c *Config) (*Logger, error) {

	c = mergeConfig(c)

	writer, err := newSyncWriter(c)
	if err != nil {
		return nil, err
	}
	return &Logger{
		Level: c.Level, // fixbug: missing level
		Log:   writer.Log,
		Logf:  writer.Logf,
		Flush: writer.Flush,
		Close: writer.Close,
	}, nil
}

func rename(path string, year int, month time.Month, day int) {

	buf := recordPool.Get().(*record)
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
	recordPool.Put(buf)

	for {
		if info, err := os.Stat(npath); info != nil || os.IsExist(err) {
			npath = npath[:nsize] + strconv.FormatInt(time.Now().UnixNano(), 36) // 重新拼过时间戳
		} else {
			if err := os.Rename(path, npath); err != nil {
				fmt.Fprintf(os.Stderr, "rename log file error: %v\n", err)
			}
			return
		}
	}

}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func Stack(all bool) []byte {
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

func Path(path string) string {
	start := strings.IndexByte(path, '$')
	if start == -1 {
		return path
	}

	buf := new(bytes.Buffer)
	mark := 0
	end := 0
	plen := len(path)
	for {
		if start == -1 {
			buf.WriteString(path[mark:])
			break
		} else {
			buf.WriteString(path[mark:start])
		}
		mark = start + 1
		if path[mark] == '{' {
			mark++
			end = nextByte(&path, '}', mark, plen)
			if end == -1 {
				buf.WriteString(path[start:])
				break
			} else {
				buf.WriteString(os.Getenv(path[mark:end]))
			}
			mark = end + 1
		} else {
			end = nextNotIdenByte(&path, mark, plen)
			if end == -1 {
				buf.WriteString(path[start:])
				break
			} else {
				buf.WriteString(os.Getenv(path[mark:end]))
			}
			mark = end
		}
		start = nextByte(&path, '$', mark, plen)
	}

	return buf.String()
}

func nextByte(v *string, c byte, start int, end int) int {
	for i := start; i < end; i++ {
		if (*v)[i] == c {
			return i
		}
	}
	return -1
}

func nextNotIdenByte(v *string, start int, end int) int {
	for i := start; i < end; i++ {
		if ch := (*v)[i]; !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
			return i
		}
	}
	return -1
}
