package gp

// LogWriter log writer interface
type LogWriter interface {
	Println(v ...interface{})
}

type nopLogger struct{}

func (nopLogger) Println(values ...interface{}) {}
