package gp

type Task struct {
	Message  any
	Handler  func(msg any) error
	Callback func(msg any, err error)
	IsRetry  func(msg any, failTimes int) bool
}

func (t *Task) reset() {
	t.Message = nil
	t.Handler = nil
	t.Callback = nil
	t.IsRetry = nil
}
