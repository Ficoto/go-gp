package gp

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	PoolCloseError    = errors.New("pool is closed")
	HandlerPanicError = errors.New("handler panic")
)

type Pool struct {
	maxPoolSize   int
	signalChannel chan struct{}
	logger        LogWriter
	isClose       bool
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
	p.signalChannel = make(chan struct{}, p.maxPoolSize)
	return p
}

func (p *Pool) safeHandler(t Task) {
	go func() {
		defer recoverPrintln(p.logger)
		defer func() {
			<-p.signalChannel
		}()
		var (
			errChan = make(chan error, 1)
			sh      = func() {
				defer func() {
					if v := recover(); v != nil {
						errChan <- fmt.Errorf("%w\ntask panic: %v\n%s\n", HandlerPanicError, v, debug.Stack())
					}
				}()
				errChan <- t.Handler(t.Message)
			}
			failTimes int
		)
		for {
			sh()
			err := <-errChan
			if t.IsRetry != nil && err != nil && t.IsRetry(t.Message, failTimes) {
				failTimes++
				continue
			}
			if t.Callback == nil {
				break
			}
			t.Callback(t.Message, err)
			break
		}
	}()
}

func (p *Pool) GoWithTask(task Task) error {
	if p.isClose {
		return PoolCloseError
	}
	p.signalChannel <- struct{}{}
	p.safeHandler(task)
	return nil
}

func (p *Pool) Go(f func()) error {
	return p.GoWithTask(Task{
		Handler: func(msg any) error {
			f()
			return nil
		},
		Callback: NopCallback,
		IsRetry:  NopIsRetry,
	})
}

func (p *Pool) Close() {
	p.isClose = true
	for len(p.signalChannel) != 0 {
		continue
	}
	close(p.signalChannel)
}

func (p *Pool) Size() int {
	return len(p.signalChannel)
}
