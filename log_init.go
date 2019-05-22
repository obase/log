package log

import (
	"github.com/obase/conf"
	"strings"
	"sync"
)

const CKEY = "logger"

var once sync.Once

func Init() {
	once.Do(func() {

		// 先初始conf
		conf.Init()

		config, ok := conf.GetMap(CKEY)
		if !ok {
			Setup(DEF_FLUSH_PERIOD, &Option{
				Level:   DEBUG,
				Path:    STDERR,
				Default: true,
			})
			return
		}

		flushPeriod, ok := conf.ElemDuration(config, "flushPeriod")
		if !ok {
			flushPeriod = DEF_FLUSH_PERIOD
		}

		options := make([]*Option, 0, 4)

		level, ok := conf.ElemString(config, "level")
		if !ok {
			level = "DEBUG"
		}
		path, ok := conf.ElemString(config, "path")
		if !ok {
			path = STDOUT
		}
		rotateBytes, ok := conf.ElemInt(config, "rotateBytes")
		rotateCycle, ok := conf.ElemString(config, "rotateCycle")
		if !ok {
			rotateCycle = "DAILY"
		}
		recordBufIdle, ok := conf.ElemInt(config, "recordBufIdle")
		if !ok {
			recordBufIdle = DEF_RECORD_BUF_IDLE //与glog相同
		}
		recordBufSize, ok := conf.ElemInt(config, "recordBufSize")
		if !ok {
			recordBufSize = DEF_RECORD_BUF_SIZE //与glog相同
		}
		writerBufSize, ok := conf.ElemInt(config, "writerBufSize")
		if !ok {
			writerBufSize = DEF_WRITER_BUF_SIZE // 与glog相同
		}

		options = append(options, &Option{
			Name:          "",
			Level:         GetLevel(level),
			Path:          path,
			RotateBytes:   int64(rotateBytes),
			RotateCycle:   GetCycle(rotateCycle),
			RecordBufIdle: recordBufIdle,
			RecordBufSize: recordBufSize,
			WriterBufSize: writerBufSize,
			Default:       true,
		})

		exts, ok := conf.ElemMap(config, "exts")
		if ok && len(exts) > 0 {
			for name, config := range exts {
				level, ok := conf.ElemString(config, "level")
				if !ok {
					level = "DEBUG"
				}
				path, ok := conf.ElemString(config, "path")
				if !ok {
					path = STDOUT
				}
				rotateBytes, ok := conf.ElemInt(config, "rotateBytes")
				rotateCycle, ok := conf.ElemString(config, "rotateCycle")
				if !ok {
					rotateCycle = "DAILY"
				}
				recordBufIdle, ok := conf.ElemInt(config, "recordBufIdle")
				if !ok {
					recordBufIdle = DEF_RECORD_BUF_IDLE //与glog相同
				}
				recordBufSize, ok := conf.ElemInt(config, "recordBufSize")
				if !ok {
					recordBufSize = DEF_RECORD_BUF_SIZE //与glog相同
				}
				writerBufSize, ok := conf.ElemInt(config, "writerBufSize")
				if !ok {
					writerBufSize = DEF_WRITER_BUF_SIZE // 与glog相同
				}

				options = append(options, &Option{
					Name:          name,
					Level:         GetLevel(level),
					Path:          path,
					RotateBytes:   int64(rotateBytes),
					RotateCycle:   GetCycle(rotateCycle),
					RecordBufIdle: recordBufIdle,
					RecordBufSize: recordBufSize,
					WriterBufSize: writerBufSize,
					Default:       false,
				})
			}
		}

		if err := Setup(flushPeriod, options...); err != nil {
			panic(err)
		}
	})
}

func GetLevel(val string) Level {
	switch strings.ToUpper(val) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	case "OFF":
		return OFF
	}
	return DEBUG
}

func GetCycle(val string) Cycle {
	switch strings.ToUpper(val) {
	case "DAILY":
		return DAILY
	case "MONTHLY":
		return MONTHLY
	case "HOURLY":
		return HOURLY
	}
	return DAILY
}
