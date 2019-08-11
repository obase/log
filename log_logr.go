package log

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Buffer struct {
	strings.Builder
	Header []byte
}

type logger struct {
	*Config
	*sync.Mutex
	*os.File
	*bufio.Writer
	*sync.Pool

	bs int64      // 已经写入字节数,用于轮转
	yr int        // 当前年份, 用于轮转
	mn time.Month // 当前月份,用于轮转
	dy int        // 当前日期,用于轮转
	hr int        // 当前小时, 用于轮转
}

func newLogger(opt *Config) (ret *logger, err error) {

	// 先判断文件是否存在, 如果存在则判断是否需要轮转
	var file *os.File
	var size int64

	var now = time.Now()
	var yr, mn, dy = now.Date()
	var hr, _, _ = now.Clock()

	if opt.Path == STDOUT {
		file = os.Stdout
	} else if opt.Path == STDERR {
		file = os.Stderr
	} else {
		if fi, fe := os.Stat(opt.Path); fi != nil || os.IsExist(fe) {

			size = fi.Size()
			buf := make([]byte, DEF_RECORD_HEADER_BYTES)
			if opt.RotateBytes > 0 && size >= opt.RotateBytes {
				if err = rename(opt.Path, yr, mn, dy, hr, buf); err != nil {
					return
				}
			} else if opt.RotateCycle != NONE {

				mtime := fi.ModTime()
				myr, mmn, mdy := mtime.Date()
				mhr, _, _ := mtime.Clock()

				switch opt.RotateCycle {
				case YEARLY:
					if myr != yr {
						if err = rename(opt.Path, yr, mn, dy, hr, buf); err != nil {
							return
						}
					}
				case MONTHLY:
					if myr != yr || mmn != mn {
						if err = rename(opt.Path, yr, mn, dy, hr, buf); err != nil {
							return
						}
					}
				case DAILY:
					if myr != yr || mmn != mn || mdy != dy {
						if err = rename(opt.Path, yr, mn, dy, hr, buf); err != nil {
							return
						}
					}
				case HOURLY:
					if myr != yr || mmn != mn || mdy != dy || mhr != hr {
						if err = rename(opt.Path, yr, mn, dy, hr, buf); err != nil {
							return
						}
					}
				}
			}
		} else {
			dir := filepath.Dir(opt.Path)
			if di, de := os.Stat(dir); di == nil && os.IsNotExist(de) {
				if err = os.MkdirAll(dir, os.ModePerm); err != nil {
					return
				}
			}
		}
		if file, err = os.OpenFile(opt.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
			return
		}
	}

	ret = &logger{
		Config: opt,
		Mutex:  new(sync.Mutex),
		File:   file,
		Writer: bufio.NewWriterSize(file, opt.WriterBufSize),
		Pool: &sync.Pool{
			New: func() interface{} {
				return &Buffer{
					Header: make([]byte, DEF_RECORD_HEADER_BYTES),
				}
			},
		},
		bs: size,
		yr: yr,
		mn: mn,
		dy: dy,
		hr: hr,
	}

	return
}

// 注意Level在外层控制避免多一重函数调用
func (lg *logger) printf(depth int, lvl Level, ctx context.Context, format string, args ...interface{}) {

	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 1
	} else if pos := strings.LastIndexByte(file, '/'); pos >= 0 {
		file = file[pos+1:]
	}

	// 获取BuffNode已经带锁, 不需要放在l.mutx范围
	buf := lg.Pool.Get().(*Buffer)

	// 生成当前时间点
	now := time.Now()
	yr, mn, dy := now.Date()         //年,月,日
	hr, mi, sc := now.Clock()        // 时,分,秒
	ms := now.Nanosecond() / 1000000 //毫秒

	// 示例: "2019-01-25 11:04:05.333 [F] ***** xxxxx.go:123 - this is a test..... "
	var idx, val int

	// 3个毫秒字符
	val = ms
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10

	// 1个点字符
	buf.Header[idx] = '.'
	idx++
	// 2个秒字符
	val = sc
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10

	// 1个冒号
	buf.Header[idx] = ':'
	idx++
	// 2个分字符
	val = mi
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	// 1个冒号
	buf.Header[idx] = ':'
	idx++
	// 2个时字符
	val = hr
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	// 1个空白
	buf.Header[idx] = SPACE
	idx++
	// 2个天字符
	val = dy
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	// 1个中横线
	buf.Header[idx] = '-'
	idx++
	// 2个月字符
	val = int(mn)
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10

	// 1个中横线
	buf.Header[idx] = '-'
	idx++
	// 4个年字符
	val = yr
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]
	idx++
	val /= 10
	buf.Header[idx] = hexs[val%10]

	// 写入buff
	for idx >= 0 {
		buf.WriteByte(buf.Header[idx])
		idx--
	}
	buf.WriteByte(SPACE)
	buf.WriteString(desc[lvl])
	buf.WriteByte(SPACE)

	buf.WriteString(file)
	buf.WriteByte(COLON)

	// 写入line,参考glog做法
	idx = 0
	for {
		buf.Header[idx] = hexs[line%10]
		if line /= 10; line <= 0 {
			break
		}
		idx++
	}
	for idx >= 0 {
		buf.WriteByte(buf.Header[idx])
		idx--
	}
	buf.WriteByte(SPACE)
	// if ctx is custom context and ctx will not be nil
	//if ctx != nil { // add trace id if provide
	//	if tid := ctx.Value(TRACEID); tid != nil {
	//		buf.WriteString(tid.(string))
	//		buf.WriteByte(SPACE)
	//	}
	//}
	buf.WriteByte(MINUS)
	buf.WriteByte(SPACE)
	// 写入msg
	fmt.Fprintf(buf, format, args...)
	if format[len(format)-1] != CRLF {
		buf.WriteByte(CRLF)
	}
	// 写到文件
	ln := buf.Len()
	// 锁定内容操作file
	lg.Mutex.Lock()
	if (lg.RotateCycle == HOURLY && (lg.yr != yr || lg.mn != mn || lg.dy != dy || lg.hr != hr)) || // 按小时
		(lg.RotateCycle == DAILY && (lg.yr != yr || lg.mn != mn || lg.dy != dy)) || // 按天
		(lg.RotateCycle == MONTHLY && (lg.yr != yr || lg.mn != mn)) || // 按月
		(lg.RotateBytes > 0 && lg.bs >= lg.RotateBytes) { // 大小
		lg.Writer.Flush()
		if lg.File != os.Stdout && lg.File != os.Stderr {
			lg.File.Close()
			rename(lg.Path, lg.yr, lg.mn, lg.dy, lg.hr, buf.Header)
			lg.File, _ = os.OpenFile(lg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			lg.Writer = bufio.NewWriter(lg.File)
		}
		lg.yr, lg.mn, lg.dy, lg.hr = yr, mn, dy, hr
		lg.bs = 0
	}
	lg.Writer.WriteString(buf.String())
	lg.bs += int64(ln)
	lg.Mutex.Unlock()

	// 归还buffer. 由于底层共用, 未写完之前不能归还
	lg.Pool.Put(buf)
}

func (lg *logger) Flush() {
	lg.Mutex.Lock()
	lg.Writer.Flush()
	lg.File.Sync()
	lg.Mutex.Unlock()
}

func (lg *logger) Close() {
	lg.Mutex.Lock()
	lg.Writer.Flush()
	lg.File.Close()
	lg.Mutex.Unlock()
}

func rename(path string, yr int, mn time.Month, dy int, hr int, tmp []byte) error {
	bd := new(strings.Builder)
	bd.WriteString(path)
	bd.WriteByte('.')

	var idx, val int

	val = hr
	for i := 0; i < 2; i++ {
		tmp[idx] = hexs[val%10]
		idx++
		val /= 10
	}
	val = dy
	for i := 0; i < 2; i++ {
		tmp[idx] = hexs[val%10]
		idx++
		val /= 10
	}
	val = int(mn)
	for i := 0; i < 2; i++ {
		tmp[idx] = hexs[val%10]
		idx++
		val /= 10
	}
	val = yr
	for i := 0; i < 4; i++ {
		tmp[idx] = hexs[val%10]
		idx++
		val /= 10
	}

	for idx > 0 {
		idx--
		bd.WriteByte(tmp[idx])
	}
	npath := bd.String()
	if fi, fe := os.Stat(npath); fi != nil || os.IsExist(fe) {
		for i := 1; true; i++ {
			bd.Reset()
			bd.WriteString(npath)
			bd.WriteByte('.')
			bd.WriteString(strconv.Itoa(i))
			xpath := bd.String()
			if fi, fe := os.Stat(xpath); fi == nil || os.IsNotExist(fe) {
				npath = xpath
				break
			}
		}
	}
	os.Rename(path, npath)
	return nil
}
