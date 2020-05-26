package log

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestReturnBuffer(t *testing.T) {

	buf := new(bytes.Buffer)
	buf.WriteString("all.log.2020-05-21")
	buf.WriteByte(DOT)
	bln := buf.Len()

	buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 36))

	path := buf.String()
	fmt.Println(path)
	fmt.Println(path[:bln])
}
