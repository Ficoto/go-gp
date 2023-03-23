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
	signalChan       chan struct{}
	internalTaskChan chan Task
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
	for _, option := range options {
		option(p)
	}
	if p.maxPoolSize == 0 {
		p.maxPoolSize = runtime.NumCPU()
	}
	if p.logger == nil {
		p.logger = nopLogger{}
	}
	return p
}

func (p *Pool) Go(f func()) error {
	return p.GoWithTask(Task{
		Handler: func(msg any) error {
			f()
			return nil
		},
	})
}

func (p *Pool) GoWithTask(task Task) error {
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

func (p *Pool) AsyncGo(f func()) {
	go func() {
		if err := p.GoWithTask(Task{
			Handler: func(msg any) error {
				f()
				return nil
			},
		}); err != nil {
			p.logger.Println("GoWithMessage fail\n", "err: ", err)
		}
	}()
}

func (p *Pool) AsyncGoWithTask(task Task) {
	go func() {
		if err := p.GoWithTask(task); err != nil {
			p.logger.Println("GoWithMessage fail\n", "err: ", err)
		}
	}()
}

func (p *Pool) Run() {
	p.signalChan = make(chan struct{}, p.maxPoolSize)
	p.taskChan = make(chan Task)
	p.internalTaskChan = make(chan Task)
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
				w := newWork(setEndSignalChan(p.signalChan), setIdleTimeout(p.idleTimeout), setLogger(p.logger), setTaskChanel(p.internalTaskChan))
				w.Run(ctx)
				p.safeWriteTask(ctx, task)
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
