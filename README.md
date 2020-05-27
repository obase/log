# package log (v2)
v2提供统一的日志代理接口, 并提供builtin的日志实现

## Installation
- go get
```
go get -u github.com/obase/log
```
- go mod 
```
go mod edit -require=github.com/obase/log@latest # v2.0.0+
```

## Configuration
```
# 系统日志设置
logger:
  # 统一刷新间隔(可选), 包括默认以及exts配置的其他日志. 默认30s.
  flushPeriod: "10s"
  # 日志级别(必需), DEBUG, INFO, ERROR, FATAL, OFF
  level: "DEBUG"
  # 日志路径(必需), stdout表示标准输出, stderr表示标准错误
  path: "/data/logs/xxx.log"
  # 轮转字节(byte)数, 默认为0表示不启用.
  #rotateBytes: 10240000
  # 轮转周期(可选), 目前支持yearly, monthly, daily, hourly
  rotateCycle: "hourly"
  # 缓冲区大小(可选), 默认256K
  bufioWriterSize: 262144
  # 其他日志. 通过key引用保存到不同的日志文件, 例如预警,追踪等场合
  exts:
    # 日志名称
    alarm:
      level:
      path: stdout
      rotateBytes:
      rotateCycle:
      bufioWriterSize:

```

## Index

### 统一日志代理

- Setup

安装代理使用的所有logger及刷新时间.
```
func Setup(flushPeriod time.Duration, g *Logger, m map[string]*Logger)
```

- Get

获取特定名称的logger, 结果可能为空
```
func Get(name string) (ret *Logger) 
```

- Must

获取特定名称的logger, 结果不能为空
```
func Must(name string) (ret *Logger)
```

- Debug

```
func Debug(args ...interface{}) 
```
- Debugf

```
func Debugf(format string, args ...interface{})
```
- Info

```
func Info(args ...interface{})
```
- Infof

```
func Infof(format string, args ...interface{})
```

- Warn

```
func Warn(args ...interface{})
```

- Warnf
```
func Warnf(format string, args ...interface{}) 
```

- Error
```
func Error(args ...interface{}) 
```

- Errorf
```
func Errorf(format string, args ...interface{})
```

- ErrorStack
```
func ErrorStack(format string, args ...interface{})
```

- Fatal
```
func Fatal(args ...interface{}) 
```

- Fatalf
```
func Fatalf(format string, args ...interface{})
```

- FatalStack
```
func FatalStack(format string, args ...interface{}) 
```

- type Logger

统一的日志结构
```
type Logger struct {
	Level Level
	Log   func(level Level, args ...interface{})
	Logf  func(level Level, format string, args ...interface{})
	Flush func()
	Close func()
}
```

- (lg *Logger) Debug
```
func (lg *Logger) Debug(args ...interface{}) 
```
- (lg *Logger) Debugf
```
func (lg *Logger) Debugf(format string, args ...interface{})
```
- (lg *Logger) Info
```
func (lg *Logger) Info(args ...interface{})
```
- (lg *Logger) Infof
```
func (lg *Logger) Infof(format string, args ...interface{})
```

- (lg *Logger) Warn
```
func (lg *Logger) Warn(args ...interface{})
```

- (lg *Logger) Warnf
```
func (lg *Logger) Warnf(format string, args ...interface{}) 
```

- (lg *Logger) Error
```
func (lg *Logger) Error(args ...interface{}) 
```

- (lg *Logger) Errorf
```
func (lg *Logger) Errorf(format string, args ...interface{})
```

- (lg *Logger) ErrorStack
```
func (lg *Logger) ErrorStack(format string, args ...interface{})
```

- (lg *Logger) Fatal
```
func (lg *Logger) Fatal(args ...interface{}) 
```

- (lg *Logger) Fatalf
```
func (lg *Logger) Fatalf(format string, args ...interface{})
```

- (lg *Logger) FatalStack
```
func (lg *Logger) FatalStack(format string, args ...interface{}) 
```


# package log (v1)
提供通用日志接口及默认实现!

## Installation
- go get
```
go get -u github.com/obase/log
```
- go mod
```
go mod edit -require=github.com/obase/log@latest
```

## Configuration
```
# 系统日志设置
logger:
  # 统一刷新间隔(可选), 包括默认以及exts配置的其他日志. 默认30s.
  flushPeriod: "10s"
  # 日志级别(必需), DEBUG, INFO, ERROR, FATAL, OFF
  level: "DEBUG"
  # 日志路径(必需), stdout表示标准输出, stderr表示标准错误
  path: "/data/logs/xxx.log"
  # 轮转字节(byte)数, 默认为0表示不启用.
  #rotateBytes: 10240000
  # 轮转周期(可选), 目前支持yearly, monthly, daily, hourly
  rotateCycle: "hourly"
  # 缓冲区大小(可选), 默认256K
  bufioWriterSize: 262144
  # 其他日志. 通过key引用保存到不同的日志文件, 例如预警,追踪等场合
  exts:
    # 日志名称
    alarm:
      level:
      path: stdout
      rotateBytes:
      rotateCycle:
      bufioWriterSize:

```

## Index
- Constants
```
const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
	OFF //不输出任何日志
)

const (
	NONE Cycle = iota
	HOURLY
	DAILY
	MONTHLY
	YEARLY
)

const (
	STDOUT = "stdout"
	STDERR = "stderr"

	DEF_FLUSH_PERIOD        = 5 * time.Second //与glog相同
	DEF_RECORD_BUF_IDLE     = 256             //与glog相同
	DEF_RECORD_BUF_SIZE     = 1024            // 默认1K
	DEF_RECORD_HEADER_BYTES = 24              // 用于header最长的栈
	DEF_WRITER_BUF_SIZE     = 256 * 1024      //与glog相同, 256k

	SPACE = '\x20'
	COLON = ':'
	MINUS = '-'
	DOT   = '.'
	CRLF  = '\n'

	TRACEID = "#tid#"
)
```
- Variables 
```

```
- type Config
```
type Level uint8
type Cycle uint8
type Config struct {
	Name            string // 日志名字
	Level           Level
	Path            string
	RotateBytes     int64
	RotateCycle     Cycle //轮转周期,目前仅支持
	BufioWriterSize int   //Buffer写缓存大小
	Default         bool  //是否默认
}

```
- func GetLog
```
func GetLog(name string) (l *logger) 
```
- func Fatal
```
var Fatal func(ctx context.Context, format string, args ...interface{})
```
- func Error
```
var Error func(ctx context.Context, format string, args ...interface{})
```
- func ErrorStack
```
var ErrorStack func(ctx context.Context, err error, all bool)
```
打印错误及堆栈信息

- func Warn
```
var Warn func(ctx context.Context, format string, args ...interface{})
```
- func Info
```
var Info func(ctx context.Context, format string, args ...interface{})
```
- func Debug
```
var Debug func(ctx context.Context, format string, args ...interface{})
```
- func Flush
```
var Flush func()
```
## Examples
```
package log

import (
	"fmt"
	//"github.com/golang/glog"
	"sync"
	"testing"
	"time"
)

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

```
ErrorStack
```
func TestLogger(t *testing.T) {
	ErrorStack(nil, "this is a error", false)
}

/*

2019-08-14 16:53:56.251 [E] logdef_test.go:12 this is a error
goroutine 6 [running]:
github.com/obase/log.stacks(0x69d900, 0x1d, 0x539, 0x45a100)
        E:/baseworkspace/src/github.com/obase/log/logimpl.go:194 +0xb8
github.com/obase/log.(*Logger).ErrorStack(0xc0000046e0, 0x0, 0x0, 0x555260, 0x5b9dc0, 0x4cdc00)
        E:/baseworkspace/src/github.com/obase/log/logimpl.go:205 +0x4d
github.com/obase/log.TestLogger(0xc0000a6200)
        E:/baseworkspace/src/github.com/obase/log/logdef_test.go:12 +0x54
testing.tRunner(0xc0000a6200, 0x599758)
        C:/Go/src/testing/testing.go:865 +0xc7
created by testing.(*T).Run
        C:/Go/src/testing/testing.go:916 +0x361

*/
```