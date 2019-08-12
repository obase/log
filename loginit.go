package log

import (
	"github.com/obase/conf"
	"strings"
)

const CKEY = "logger"

func init() {
	config, ok := conf.GetMap(CKEY)
	if !ok {
		Setup(DEF_FLUSH_PERIOD, &Config{
			Level:   DEBUG,
			Path:    STDERR,
			Default: true,
		})
	} else {
		flushPeriod, ok := conf.ElemDuration(config, "flushPeriod")

		options := make([]*Config, 0, 4)

		level, _ := conf.ElemString(config, "level")
		path, _ := conf.ElemString(config, "path")
		rotateBytes, _ := conf.ElemInt(config, "rotateBytes")
		rotateCycle, _ := conf.ElemString(config, "rotateCycle")
		bufioWriterSize, _ := conf.ElemInt(config, "bufioWriterSize")
		options = append(options, &Config{
			Name:            "",
			Level:           GetLevel(level),
			Path:            path,
			RotateBytes:     int64(rotateBytes),
			RotateCycle:     GetCycle(rotateCycle),
			BufioWriterSize: bufioWriterSize,
			Default:         true,
		})

		exts, ok := conf.ElemMap(config, "exts")
		if ok && len(exts) > 0 {
			for name, config := range exts {
				level, _ := conf.ElemString(config, "level")
				path, _ := conf.ElemString(config, "path")
				rotateBytes, _ := conf.ElemInt(config, "rotateBytes")
				rotateCycle, _ := conf.ElemString(config, "rotateCycle")
				bufioWriterSize, _ := conf.ElemInt(config, "bufioWriterSize")
				options = append(options, &Config{
					Name:            name,
					Level:           GetLevel(level),
					Path:            path,
					RotateBytes:     int64(rotateBytes),
					RotateCycle:     GetCycle(rotateCycle),
					BufioWriterSize: bufioWriterSize,
					Default:         false,
				})
			}
		}
		if err := Setup(flushPeriod, options...); err != nil {
			panic(err)
		}
	}

	// 初始化全局定义
	Debug, Info, Warn, Error, ErrorStack, Fatal, Flush = _default.Debug, _default.Info, _default.Warn, _default.Error, _default.ErrorStack, _default.Fatal, FlushAll
	Debugf, Infof, Warnf, Errorf, Fatalf, Flushf = _default.Debug, _default.Info, _default.Warn, _default.Error, _default.Fatal, FlushAll
	Inforf = _default.Info
}

func FlushAll() {
	if _default != nil {
		_default.Flush()
	}
	for _, v := range _loggers {
		v.Flush()
	}
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
