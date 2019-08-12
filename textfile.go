package log

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type TextfileWriter struct {
	*Config
	*os.File
	*bufio.Writer
	Year  int
	Month time.Month
	Day   int
	Hour  int
	Size  int64
}

func NewTextfileWriter(c *Config) (Writer, error) {
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
	if file, err = os.OpenFile(c.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		return nil, err
	}
	return &TextfileWriter{
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
func rename(path string, year int, month time.Month, day int, hour int) error {

	buf := recordSyncPool.Get().(*Record)
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

func (w *TextfileWriter) Write(r *Record) {
	size := int64(r.Len())

	rotated := false
	if w.Config.RotateBytes > 0 && w.Config.RotateBytes < w.Size+size {
		rotated = true
	} else {
		switch w.Config.RotateCycle {
		case NONE:
			rotated = false
		case YEARLY:
			rotated = (w.Year != r.Year)
		case MONTHLY:
			rotated = (w.Year != r.Year) || (w.Month != r.Month)
		case DAILY:
			rotated = (w.Year != r.Year) || (w.Month != r.Month) || (w.Day != r.Day)
		case HOURLY:
			rotated = (w.Year != r.Year) || (w.Month != r.Month) || (w.Day != r.Day) || (w.Hour != r.Hour)
		}
	}
	if rotated {
		w.Writer.Flush() //刷新缓存
		w.File.Close()   // 关闭句柄
		rename(w.Config.Path, w.Year, w.Month, w.Day, w.Hour)
		w.File, _ = os.OpenFile(w.Config.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		w.Writer = bufio.NewWriterSize(w.File, w.Config.BufioWriterSize)
		w.Year, w.Month, w.Day, w.Hour, w.Size = r.Year, r.Month, r.Day, r.Hour, 0
	}
	w.Writer.Write(r.Bytes())
}

func (w *TextfileWriter) Flush() {
	w.Writer.Flush()
	w.File.Sync()
}

func (w *TextfileWriter) Close() {
	w.Writer.Flush()
	w.File.Close()
}
