package gp

import "runtime/debug"

// LogWriter log writer interface
type LogWriter interface {
	Println(values ...any)
	Printf(format string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Println(values ...any) {}

func (nopLogger) Printf(format string, args ...any) {}

func recoverPrintln(log LogWriter) {
	if v := recover(); v != nil {
		log.Printf("task panic: %v\n%s\n", v, debug.Stack())
	}
}
