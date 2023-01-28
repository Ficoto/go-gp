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
)

type work struct {
	taskChan       chan Task
	callbackChan   chan callbackMessage
	currentTask    Task
	signalChan     chan struct{}
	idleTimeout    time.Duration
	lastFinishTime time.Time
	logger         LogWriter
}

type workOption func(w *work)

func setCallbackChan(callbackChan chan callbackMessage) workOption {
	return func(w *work) {
		w.callbackChan = callbackChan
	}
}

func setIdleTimeout(idleTimeout time.Duration) workOption {
	return func(w *work) {
		w.idleTimeout = idleTimeout
	}
}

func setSignalChan(sc chan struct{}) workOption {
	return func(w *work) {
		w.signalChan = sc
	}
}

func setLogger(lw LogWriter) workOption {
	return func(w *work) {
		w.logger = lw
	}
}

func newWork(tc chan Task, options ...workOption) *work {
	var w = new(work)
	w.taskChan = tc
	w.logger = nopLogger{}
	for _, option := range options {
		option(w)
	}
	if w.idleTimeout == 0 {
		w.idleTimeout = defaultIdleTimeOut
	}
	return w
}

func (w *work) Run(ctx context.Context) {
	go func() {
		defer func() {
			if v := recover(); v != nil {
				w.logger.Println(v, "\n", string(debug.Stack()))
				w.handlerCallback(HandlerPanicError)
			}
			if w.signalChan == nil {
				return
			}
			w.signalChan <- struct{}{}
		}()
		tc := time.Tick(w.idleTimeout / 2)
		var ok bool
		for {
			select {
			case w.currentTask, ok = <-w.taskChan:
				if !ok {
					return
				}
				err := w.currentTask.Handler(w.currentTask.Message)
				w.handlerCallback(err)
				w.lastFinishTime = time.Now()
			case <-tc:
				if w.lastFinishTime.Add(w.idleTimeout).Before(time.Now()) {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (w *work) handlerCallback(err error) {
	if w.callbackChan == nil {
		return
	}
	w.callbackChan <- callbackMessage{
		task: w.currentTask,
		err:  err,
	}
}
