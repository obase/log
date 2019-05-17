package log

import (
	. "context"
	"fmt"
	"testing"
	"time"
)

var ctx Context = WithValue(TODO(), TRACEID, RandTraceId())

func TestZlogPerf(t *testing.T) {
	defer Flush()
	start := time.Now()
	//for i := 0; i < 1; i++ {
	Debug(ctx, "this is Debug on %v\n", time.Now())
	//	Info(ctx, "this is Info on %v", time.Now())
	//	Warn(ctx, "this is Warn on %v", time.Now())
	//	Error(ctx, "this is Error on %v\n", time.Now())
	//	Fatal(ctx, "this is Fatal on %v", time.Now())
	//}
	end := time.Now()
	fmt.Println("used: ", end.Sub(start).Nanoseconds())
}

//func TestGlogPerf(t *testing.T) {
//	defer glog.Flush()
//	flag.Parse()
//	flag.Set("log_dir", "e:/Temp/")
//	flag.Set("logtostderr", "false")
//	flag.Set("stderrthreshold", "FATAL")
//	start := time.Now()
//	for i := 0; i < 1000000; i++ {
//		glog.Infof("this is Info on %v", time.Now())
//		glog.Infof("this is Info on %v", time.Now())
//		glog.Infof("this is Info on %v", time.Now())
//		glog.Errorf("this is Error on %v", time.Now())
//		glog.Errorf("this is Error on %v", time.Now())
//	}
//	end := time.Now()
//	fmt.Println("used: ", end.Sub(start).Nanoseconds())
//}

func TestRotateBytes(t *testing.T) {
	fmt.Printf("test")
}
