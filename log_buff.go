package log

import (
	"bytes"
	"sync"
)

type Buffer struct {
	*bytes.Buffer
	Tmp []byte
	Nxt *Buffer
}

type RecordBuffPool struct {
	Mux  *sync.Mutex
	Max  int
	Len  int
	Head *Buffer
	Init int
}

func NewBuffPool(max, init int) *RecordBuffPool {
	return &RecordBuffPool{
		Mux:  new(sync.Mutex),
		Max:  max,
		Len:  0,
		Head: nil,
		Init: init,
	}
}

func (bp *RecordBuffPool) Get() (bf *Buffer) {
	bp.Mux.Lock()
	if bp.Head != nil {
		bf = bp.Head
		bp.Head = bf.Nxt
		bp.Len--
		bp.Mux.Unlock()
		bf.Reset()
		return
	}
	bp.Mux.Unlock()

	bf = &Buffer{
		Buffer: bytes.NewBuffer(make([]byte, 0, bp.Init)), // 务必设置len为0
		Nxt:    nil,
		Tmp:    make([]byte, DEF_RECORD_HEADER_BYTES),
	}
	return
}

func (bp *RecordBuffPool) Put(bf *Buffer) {
	if bp.Len >= bp.Max {
		return
	}
	bp.Mux.Lock()
	bf.Nxt = bp.Head
	bp.Head = bf
	bp.Len++
	bp.Mux.Unlock()
}
