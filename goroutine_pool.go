package gp

import (
	"context"
	"errors"
	"runtime"
	"runtime/debug"
	"time"
)

var PoolCloseError = errors.New("pool is closed")

type Pool struct {
	taskChan         chan Task
	idleTimeout      time.Duration
	maxPoolSize      int
	callback         func(task Task, err error)
	signalChan       chan struct{}
	internalTaskChan chan Task
	callbackChan     chan callbackMessage
	ctxCancel        context.CancelFunc
	isClose          bool
	logger           LogWriter
}

type Option func(p *Pool)

func SetMaxPoolSize(mps int) Option {
	return func(p *Pool) {
		if mps <= 0 {
			return
		}
		p.maxPoolSize = mps
	}
}

func SetCallback(handler func(task Task, err error)) Option {
	return func(p *Pool) {
		p.callback = handler
	}
}

func SetErrorCallback(handler func(task Task, err error)) Option {
	return func(p *Pool) {
		p.callback = func(task Task, err error) {
			if err == nil {
				return
			}
			handler(task, err)
		}
	}
}

func SetIdleTimeout(idleTimeOut time.Duration) Option {
	return func(p *Pool) {
		p.idleTimeout = idleTimeOut
	}
}

func SetLogger(lw LogWriter) Option {
	return func(p *Pool) {
		p.logger = lw
	}
}

func New(options ...Option) *Pool {
	var p = new(Pool)
	p.maxPoolSize = runtime.NumCPU()
	p.logger = nopLogger{}
	for _, option := range options {
		option(p)
	}
	return p
}

func (p *Pool) Go(f func() error) error {
	return p.GoWithMessage(Task{
		Message: nil,
		Handler: func(msg interface{}) error {
			return f()
		},
	})
}

func (p *Pool) GoWithMessage(task Task) error {
	if p.isClose {
		return PoolCloseError
	}
	defer func() {
		if v := recover(); v != nil {
			p.logger.Println(v, "\n", string(debug.Stack()))
		}
	}()
	p.taskChan <- task
	return nil
}

func (p *Pool) AsyncGo(f func() error) {
	go func() {
		if err := p.GoWithMessage(Task{
			Message: nil,
			Handler: func(msg interface{}) error {
				return f()
			},
		}); err != nil {
			p.logger.Println("GoWithMessage fail\n", "err: ", err)
		}
	}()
}

func (p *Pool) AsyncGoWithMessage(task Task) {
	go func() {
		if err := p.GoWithMessage(task); err != nil {
			p.logger.Println("GoWithMessage fail\n", "err: ", err)
		}
	}()
}

func (p *Pool) Run() {
	p.signalChan = make(chan struct{}, p.maxPoolSize)
	p.taskChan = make(chan Task)
	p.internalTaskChan = make(chan Task)
	if p.callback != nil {
		p.callbackChan = make(chan callbackMessage)
	}
	var ctx = context.Background()
	ctx, p.ctxCancel = context.WithCancel(ctx)
	for i := 0; i < p.maxPoolSize; i++ {
		p.signalChan <- struct{}{}
	}
	go func() {
		for {
			select {
			case task, ok := <-p.taskChan:
				if !ok {
					return
				}
				if len(p.signalChan) == 0 {
					p.safeWriteTask(ctx, task)
					continue
				}
				_, ok = <-p.signalChan
				if !ok {
					return
				}
				w := newWork(p.internalTaskChan, setSignalChan(p.signalChan), setCallbackChan(p.callbackChan), setIdleTimeout(p.idleTimeout), setLogger(p.logger))
				w.Run(ctx)
				p.safeWriteTask(ctx, task)
			}
		}
	}()
	go func() {
		if p.callback == nil {
			return
		}
		for {
			select {
			case cMsg, ok := <-p.callbackChan:
				if !ok {
					return
				}
				p.safeHandleCallback(cMsg)
			}
		}
	}()
}

func (p *Pool) Close() {
	p.isClose = true
	close(p.taskChan)
	p.ctxCancel()
	for {
		if len(p.signalChan) == cap(p.signalChan) {
			break
		}
	}
	close(p.signalChan)
	close(p.internalTaskChan)
	if p.callback != nil {
		close(p.callbackChan)
	}
}

func (p *Pool) Size() int {
	return p.maxPoolSize - len(p.signalChan)
}

func (p *Pool) safeWriteTask(ctx context.Context, task Task) {
	defer func() {
		if v := recover(); v != nil {
			p.logger.Println(v, "\n", string(debug.Stack()))
		}
	}()
	select {
	case <-ctx.Done():
		return
	default:
		p.internalTaskChan <- task
	}
}

func (p *Pool) safeHandleCallback(cMsg callbackMessage) {
	defer func() {
		if v := recover(); v != nil {
			p.logger.Println(v, "\n", string(debug.Stack()))
		}
	}()
	p.callback(cMsg.task, cMsg.err)
}
