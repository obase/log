package log

import (
	"fmt"
	"testing"
	"time"
)

func TestGetLog(t *testing.T) {
	for {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
		Info(nil, "this is a test")
		Flush()
		time.Sleep(10 * time.Second)
	}

}
