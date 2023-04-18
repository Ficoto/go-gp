package gp

type Task struct {
	Message  any
	Handler  func(msg any) error
	Callback func(msg any, err error)
	IsRetry  func(msg any, failTimes int) bool
}

func NopCallback(msg any, err error) {}

func NopIsRetry(msg any, failTimes int) bool { return false }
