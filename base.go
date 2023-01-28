package gp

type Task struct {
	Message interface{}
	Handler func(msg interface{}) error
}

type callbackMessage struct {
	task Task
	err  error
}
