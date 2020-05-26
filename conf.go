package log

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Level uint

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
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
	Name            string // 日志名字
	Level           Level
	Path            string
	RotateBytes     int64
	RotateCycle     Cycle //轮转周期,目前仅支持
	BufioWriterSize int   //Buffer写缓存大小
	Async           bool  // 是否使用异步
	AsynChanSize    int   // 异步管道容量,默认512
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
	if c.AsynChanSize == 0 {
		c.AsynChanSize = 512
	}
	return c
}

const (
	DOT   = '.'
	MINUS = '-'
)

var desc [5]byte = [5]byte{'D', 'I', 'W', 'E', 'F'}
var hexs [16]byte = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func rename(path string, year int, month time.Month, day int) {

	buf := BorrowBuffer()
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
	ReturnBuffer(buf)

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
