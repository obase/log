package zlog

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
		return
	}

	flushPeriod, ok := conf.ElemDuration(config, "flushPeriod")
	if !ok {
		flushPeriod = DEF_FLUSH_PERIOD
	}

	options := make([]*Config, 0, 4)

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
	bufioWriterSize, ok := conf.ElemInt(config, "bufioWriterSize")
	if !ok {
		bufioWriterSize = DEF_BUFIO_WRITER_SIZE // 与glog相同
	}

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
			bufioWriterSize, ok := conf.ElemInt(config, "bufioWriterSize")
			if !ok {
				bufioWriterSize = DEF_BUFIO_WRITER_SIZE // 与glog相同
			}

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

func mergeConfig(c *Config) *Config {
	if c == nil {
		c = new(Config)
	}
	if c.BufioWriterSize <= 0 {
		c.BufioWriterSize = DEF_BUFIO_WRITER_SIZE
	}
	return c
}
