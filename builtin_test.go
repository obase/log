package log

import (
	"testing"
)

func TestReturnBuffer(t *testing.T) {
	defer Close()
	Debug("this", "is", "a", "debug")
	Info("this", "is", "a", "warn")
	Warn("this", "is", "a", "warn")
	Error("this", "is", "a", "error")
	//Fatal("this", "is", "a", "fatal")

	notify := Get("notify")
	notify.Debug("this", "is", "a", "debug@notify")
	notify.Info("this", "is", "a", "warn@notify")
	notify.Warn("this", "is", "a", "warn@notify")
	notify.Error("this", "is", "a", "error@notify")
}
