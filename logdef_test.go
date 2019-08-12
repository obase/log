package log

import (
	"errors"
	//"github.com/golang/glog"
	"sync"
	"testing"
)

func TestGetLog(t *testing.T) {
	err := errors.New("this is a test")
	ErrorStack(nil, err, false)
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
