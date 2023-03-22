package gp

import (
	"context"
	"errors"
	"runtime/debug"
	"time"
)

const (
	defaultIdleTimeOut = time.Minute * 10
)

var (
	HandlerPanicError = errors.New("handler panic")
	HandlerNilError   = errors.New("handler is nil")
)

type work struct {
	taskChan       chan Task
	isWorkTC       bool
	currentTask    Task
	endSignalChan  chan struct{}
	idleTimeout    time.Duration
	lastFinishTime time.Time
	logger         LogWriter
	errorChan      chan error
}

type workOption func(w *work)

func setIdleTimeout(idleTimeout time.Duration) workOption {
	return func(w *work) {
		w.idleTimeout = idleTimeout
	}
}

func setEndSignalChan(sc chan struct{}) workOption {
	return func(w *work) {
		w.endSignalChan = sc
	}
}

func setLogger(lw LogWriter) workOption {
	return func(w *work) {
		w.logger = lw
	}
}

func setTaskChanel(tc chan Task) workOption {
	return func(w *work) {
		w.taskChan = tc
	}
}

func newWork(options ...workOption) *work {
	var w = new(work)
	for _, option := range options {
		option(w)
	}
	if w.idleTimeout == 0 {
		w.idleTimeout = defaultIdleTimeOut
	}
	if w.taskChan == nil {
		w.taskChan = make(chan Task)
		w.isWorkTC = true
	}
	if w.logger == nil {
		w.logger = nopLogger{}
	}
	w.errorChan = make(chan error, 1)
	return w
}

func (w *work) Run(ctx context.Context) {
	go func() {
		tc := time.Tick(w.idleTimeout / 2)
		var ok bool
		for {
			select {
			case w.currentTask, ok = <-w.taskChan:
				if !ok {
					break
				}
				w.safeHandler()
				err := <-w.errorChan
				if err != nil && w.currentTask.IsRetry != nil {
					for i := 1; w.currentTask.IsRetry(w.currentTask.Message, i); i++ {
						w.safeHandler()
						err = <-w.errorChan
						if err == nil {
							break
						}
					}
				}
				if w.currentTask.Callback != nil {
					w.safeCallback(err)
				}
				w.currentTask.reset()
				w.lastFinishTime = time.Now()
			case <-tc:
				if w.lastFinishTime.Add(w.idleTimeout).Before(time.Now()) {
					ok = false
				}
			case <-ctx.Done():
				ok = false
			}
			if !ok {
				break
			}
		}
		close(w.errorChan)
		if w.isWorkTC {
			close(w.taskChan)
		}
		if w.endSignalChan == nil {
			return
		}
		w.endSignalChan <- struct{}{}
	}()
}

func (w *work) InputTask() chan Task {
	return w.taskChan
}

func (w *work) safeCallback(err error) {
	if w.currentTask.Callback == nil {
		return
	}
	defer func() {
		if v := recover(); v != nil {
			w.logger.Println(v, "\n", string(debug.Stack()))
		}
	}()
	w.currentTask.Callback(w.currentTask.Message, err)
}

func (w *work) safeHandler() {
	defer func() {
		if v := recover(); v != nil {
			w.logger.Println(v, "\n", string(debug.Stack()))
			w.errorChan <- HandlerPanicError
		}
	}()
	if w.currentTask.Handler == nil {
		w.errorChan <- HandlerNilError
		return
	}
	w.errorChan <- w.currentTask.Handler(w.currentTask.Message)
}
