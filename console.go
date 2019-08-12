package log

import "os"

type ConsoleWriter struct {
	File *os.File
}

func NewConsoleWriter(f *os.File) Writer {
	return &ConsoleWriter{
		File: f,
	}
}

func (w *ConsoleWriter) Write(r *Record) {
	w.File.Write(r.Bytes())
}

func (w *ConsoleWriter) Flush() {
	// nothing
}

func (w *ConsoleWriter) Close() {
	// nothing
}
