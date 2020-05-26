package log

import (
	"bytes"
	"sync"
	"time"
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
