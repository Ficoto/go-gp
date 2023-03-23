package gp

// LogWriter log writer interface
type LogWriter interface {
	Println(values ...any)
}

type nopLogger struct{}

func (nopLogger) Println(values ...any) {}
