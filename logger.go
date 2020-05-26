package log

type Logger struct {
	Level  Level
	Writer interface{}
	Log    func(level Level, args ...interface{})
	Logf   func(level Level, format string, args ...interface{})
	Logs   func(level Level, format string, args ...interface{})
	Flush  func()
	Close  func()
}

func NewLogger(c *Config) (ret *Logger, err error) {
	if c.Async {
		//writer, err := NewAsyncWriter(c)
		if err != nil {
			return
		}
		//ret = &Logger{
		//	Writer: writer,
		//	Write:  writer.Write,
		//	Writef: writer.Writef,
		//	Flush:  writer.Flush,
		//	Close:  writer.Close,
		//}
	} else {
		writer, err := NewSyncWriter(c)
		if err != nil {
			return
		}
		ret = &Logger{
			Writer: writer,
			Log:    writer.Log,
			Logf:   writer.Logf,
			Logs:   writer.Logs,
			Flush:  writer.Flush,
			Close:  writer.Close,
		}
	}
	return
}

func (lg *Logger) Debug(args ...interface{}) {
	if DEBUG >= lg.Level {
		lg.Log(DEBUG, args...)
	}
}

func (lg *Logger) Debugf(format string, args ...interface{}) {
	if DEBUG >= lg.Level {
		lg.Logf(DEBUG, format, args...)
	}
}

func (lg *Logger) Info(args ...interface{}) {
	if INFO >= lg.Level {
		lg.Log(INFO, args...)
	}
}

func (lg *Logger) Infof(format string, args ...interface{}) {
	if INFO >= lg.Level {
		lg.Logf(INFO, format, args...)
	}
}

func (lg *Logger) Warn(args ...interface{}) {
	if WARN >= lg.Level {
		lg.Log(WARN, args...)
	}
}

func (lg *Logger) Warnf(format string, args ...interface{}) {
	if WARN >= lg.Level {
		lg.Logf(WARN, format, args...)
	}
}

func (lg *Logger) Error(args ...interface{}) {
	if ERROR >= lg.Level {
		lg.Log(ERROR, args...)
	}
}

func (lg *Logger) Errorf(format string, args ...interface{}) {
	if ERROR >= lg.Level {
		lg.Logf(ERROR, format, args...)
	}
}

func (lg *Logger) Fatal(args ...interface{}) {
	if FATAL >= lg.Level {
		lg.Log(FATAL, args...)
	}
}

func (lg *Logger) Fatalf(format string, args ...interface{}) {
	if FATAL >= lg.Level {
		lg.Logf(FATAL, format, args...)
	}
}
