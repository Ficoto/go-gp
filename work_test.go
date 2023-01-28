package gp

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestWork_Run(t *testing.T) {
	w1 := newWork(nil)
	w1.Run(context.Background())

	f := func(msg interface{}) error {
		t.Log(msg)
		return nil
	}

	taskChan1 := make(chan Task)
	w2 := newWork(taskChan1)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	w2.Run(ctx)
	go func() {
		time.Sleep(time.Second * 10)
		t.Log(fmt.Sprintf("this is testing goroutine %v", <-taskChan1))
	}()
	taskChan1 <- Task{
		Message: "test before cancel",
		Handler: f,
	}
	cancel()
	taskChan1 <- Task{
		Message: "test after cancel",
		Handler: f,
	}

	taskChan2 := make(chan Task)
	w3 := newWork(taskChan2, setIdleTimeout(time.Second*10))
	w3.Run(context.Background())
	taskChan2 <- Task{
		Message: "test1",
		Handler: f,
	}
	time.Sleep(time.Second * 6)
	taskChan2 <- Task{
		Message: "test2",
		Handler: f,
	}
	time.Sleep(time.Second * 15)
	go func() {
		t.Logf("this is testing goroutine %v", <-taskChan2)
	}()
	taskChan2 <- Task{
		Message: "test3",
		Handler: f,
	}
}

func TestWork_RunWithSignalChan(t *testing.T) {
	//taskChan1 := make(chan Task)
	//signalChan1 := make(chan struct{})
	//w1 := newWork(taskChan1, setSignalChan(signalChan1))
	//w1.Run(context.Background())
	//taskChan1 <- Task{
	//	Message: "test",
	//	Handler: func(msg interface{}) error {
	//		panic("test")
	//	},
	//}
	//t.Logf("work is end %v", <-signalChan1)

	taskChan2 := make(chan Task)
	signalChan2 := make(chan struct{})
	w2 := newWork(taskChan2, setSignalChan(signalChan2))
	ctx, cancel := context.WithCancel(context.Background())
	w2.Run(ctx)
	taskChan2 <- Task{
		Message: "test",
		Handler: func(msg interface{}) error {
			return nil
		},
	}
	cancel()
	t.Logf("work is end %v", <-signalChan2)
}

func TestWork_RunWithErrChan(t *testing.T) {
	taskChan := make(chan Task)
	callbackChan := make(chan callbackMessage)
	w := newWork(taskChan, setCallbackChan(callbackChan))
	w.Run(context.Background())
	go func() {
		for {
			select {
			case errMsg := <-callbackChan:
				t.Log(errMsg)
			}
		}
	}()
	for i := 0; i != 10; i++ {
		go func(i int) {
			taskChan <- Task{
				Message: fmt.Sprintf("test %d", i),
				Handler: func(msg interface{}) error {
					return errors.New(fmt.Sprintf("error %d", i))
				},
			}
		}(i)
	}
	time.Sleep(time.Second * 3)
}
