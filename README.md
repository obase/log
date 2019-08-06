# package log
提供通用日志接口及默认实现!

# Installation
- go get
```
go get -u github.com/obase/log
```
- go mod
```
go mod edit -require=github.com/obase/log@latest
```

# Configuration
```
# 系统日志设置
logger:
  # 统一刷新间隔(可选), 包括默认以及exts配置的其他日志. 默认30s.
  flushPeriod: "5s"

  # 日志级别(必需), DEBUG, INFO, ERROR, FATAL, OFF
  level: "DEBUG"
  # 日志路径(必需), stdout表示标准输出, stderr表示标准错误
  path: "log/xxx.log"
  # 轮转字节(byte)数, 默认为0表示不启用.
  rotateBytes: 10240000
  # 轮转周期(可选), 目前支持yearly, monthly, daily, hourly
  rotateCycle: "daily"
  # 记录缓存池空闲(可选), 默认256
  recordBufIdle: 256
  # 记录缓存区大小(可选), 默认1024字节
  recordBufSize: 1024
  # 缓冲区大小(可选), 默认256K
  writerBufSize: 262144

  # 其他日志. 通过key引用保存到不同的日志文件, 例如预警,追踪等场合
  exts:
    # 日志名称
    alarm:
      level:
      path: stdout
      rotateBytes:
      rotateCycle:
      recordBufIdle:
      recordBufSize:
      writerBufSize
```

# Index
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
	Name          string `json:"name" bson:"name" yaml:"name"` // 日志名字
	Level         Level  `json:"level" bson:"level" yaml:"level"`
	Path          string `json:"path" bson:"path" yaml:"path"`
	RotateBytes   int64  `json:"rotateBytes" bson:"rotateBytes" yaml:"rotateBytes"`
	RotateCycle   Cycle  `json:"rotateCycle" bson:"rotateCycle" yaml:"rotateCycle"`       //轮转周期,目前仅支持
	RecordBufIdle int    `json:"recordBufIdle" bson:"recordBufIdle" yaml:"recordBufIdle"` // record buf的空闲数量
	RecordBufSize int    `json:"recordBufSize" bson:"recordBufSize" yaml:"recordBufSize"` // record buf的初始大小
	WriterBufSize int    `json:"writerBufSize" bson:"writerBufSize" yaml:"writerBufSize"` //Buffer写缓存大小
	Default       bool   `json:"default" bson:"default" yaml:"default"`                   //是否默认
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
# Examples
```
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
```