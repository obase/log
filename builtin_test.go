package log

import (
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGetLog(t *testing.T) {
	defer _glog.Flush()
	//defer Flush()
	flag.Set("log_dir", `E:\data\logs`)
	flag.Parse()
	paral := 100
	times := 100 * 10000
	start := time.Now().UnixNano()
	testInfo(paral, times)
	end := time.Now().UnixNano()
	fmt.Println("used (ms):", (end-start)/1000000)
}

func testInfo(paral int, times int) {
	wg := sync.WaitGroup{}
	for j := 0; j < paral; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < times; i++ {
				//Info(nil, "this is a test, j=%v, i=%v", j, i)
				_glog.Infof("this is a test, j=%v, i=%v", j, i)
			}
		}()
	}
	wg.Wait()
}

func TestDebug(t *testing.T) {
	Debug("this", "is")
}
