package gp

type Task struct {
	Message  interface{}
	Handler  func(msg interface{}) error
	Callback func(msg interface{}, err error)
	IsRetry  func(msg interface{}, failTimes int) bool
}

func (t *Task) reset() {
	t.Message = nil
	t.Handler = nil
	t.Callback = nil
	t.IsRetry = nil
}
