package log

import (
	"fmt"
	//"github.com/golang/glog"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	ErrorStack(nil, "this is a error", false)
}

func TestGetLog(t *testing.T) {
	//defer glog.Flush()
	defer Flush()
	//flag.Set("log_dir", `E:\data\logs`)
	//flag.Parse()
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
				Info(nil, "this is a test, j=%v, i=%v", j, i)
				//glog.Infof("this is a test, j=%v, i=%v", j, i)
			}
		}()
	}
	wg.Wait()
}

/*

与glog的性能对比:
glog
used (ms): 62551
used (ms): 59141
used (ms): 65945

zlog:
used (ms): 33264
used (ms): 32110
used (ms): 41861
*/
