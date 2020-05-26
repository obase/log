package log

import (
	"github.com/obase/conf"
	"strings"
	"time"
)

const CKEY = "logger"

func init() {
	config, ok := conf.GetMap(CKEY)
	if !ok {
		stdout, err := newBuiltinLogger(&Config{
			Level: DEBUG,
			Path:  STDERR,
		})
		if err != nil {
			panic(err)
		}
		Setup(DEFAULT_FLUSH_PERIOD, stdout, nil)
	} else {

		var (
			global      *Logger
			logmap      = make(map[string]*Logger)
			flushPeriod time.Duration

			level           string
			path            string
			rotateBytes     int64
			rotateCycle     string
			bufioWriterSize int
			async           bool
			asyncLimit      int
			err             error
		)

		flushPeriod, _ = conf.ElemDuration(config, "flushPeriod")

		level, _ = conf.ElemString(config, "level")
		path, _ = conf.ElemString(config, "path")
		rotateBytes, _ = conf.ElemInt64(config, "rotateBytes")
		rotateCycle, _ = conf.ElemString(config, "rotateCycle")
		bufioWriterSize, _ = conf.ElemInt(config, "bufioWriterSize")
		async, _ = conf.ElemBool(config, "async")
		asyncLimit, _ = conf.ElemInt(config, "asyncLimit")

		global, err = newBuiltinLogger(&Config{
			Level:           GetLevel(level),
			Path:            path,
			RotateBytes:     rotateBytes,
			RotateCycle:     GetCycle(rotateCycle),
			BufioWriterSize: bufioWriterSize,
			Async:           async,
			AsyncLimit:      asyncLimit,
		})
		if err != nil {
			panic(err)
		}

		exts, _ := conf.ElemMap(config, "exts")
		for name, config := range exts {
			level, _ = conf.ElemString(config, "level")
			path, _ = conf.ElemString(config, "path")
			rotateBytes, _ = conf.ElemInt64(config, "rotateBytes")
			rotateCycle, _ = conf.ElemString(config, "rotateCycle")
			bufioWriterSize, _ = conf.ElemInt(config, "bufioWriterSize")
			async, _ = conf.ElemBool(config, "async")
			asyncLimit, _ = conf.ElemInt(config, "asyncLimit")

			var logger *Logger
			logger, err = newBuiltinLogger(&Config{
				Level:           GetLevel(level),
				Path:            path,
				RotateBytes:     rotateBytes,
				RotateCycle:     GetCycle(rotateCycle),
				BufioWriterSize: bufioWriterSize,
				Async:           async,
				AsyncLimit:      asyncLimit,
			})
			if err != nil {
				panic(err)
			}
			logmap[name] = logger
		}

		Setup(flushPeriod, global, logmap)

	}
}

func GetLevel(v string) Level {
	switch strings.ToUpper(v) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return DEBUG
	}
}

func GetCycle(v string) Cycle {
	switch strings.ToUpper(v) {
	case "DAILY":
		return DAILY
	case "MONTHLY":
		return MONTHLY
	case "YEARLY":
		return YEARLY
	default:
		return NEVER
	}
}
