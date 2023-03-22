package gp

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestWork_Run(t *testing.T) {
	w := newWork()
	w.Run(context.Background())
	for i := 0; i != 10; i++ {
		go func(i int) {
			w.InputTask() <- Task{
				Message: fmt.Sprintf("test %d", i),
				Handler: func(msg interface{}) error {
					return errors.New(fmt.Sprintf("error %d", i))
				},
				Callback: func(msg interface{}, err error) {
					t.Log("task finish,err:", err)
				},
				IsRetry: func(msg interface{}, failTimes int) bool {
					if failTimes <= 3 {
						return true
					}
					t.Log("fail times: ", failTimes)
					return false
				},
			}
		}(i)
	}
	time.Sleep(time.Second)
}

func TestSafeHandler(t *testing.T) {
	var w = newWork()
	w.Run(context.Background())
	w.InputTask() <- Task{
		Message: "test task",
		Handler: func(msg interface{}) error {
			panic(msg)
		},
		Callback: func(msg interface{}, err error) {
			t.Log(msg, err)
		},
	}
	time.Sleep(time.Second)
}

func TestSafeCallback(t *testing.T) {
	var w = newWork()
	w.Run(context.Background())
	w.InputTask() <- Task{
		Message: "test task",
		Handler: func(msg interface{}) error {
			t.Log("this is handler")
			return nil
		},
		Callback: func(msg interface{}, err error) {
			t.Log("this is callback")
			panic(err)
		},
	}
	time.Sleep(time.Second)
}

func TestSetIdleTimeout(t *testing.T) {
	var sc1 = make(chan struct{})
	var w1 = newWork(setIdleTimeout(time.Second), setEndSignalChan(sc1))
	w1.Run(context.Background())
	var since1 = time.Now()
	w1.InputTask() <- Task{
		Message: since1,
		Handler: func(msg interface{}) error {
			ti, _ := msg.(time.Time)
			t.Log(time.Now().Sub(ti))
			return nil
		},
	}
	<-sc1
	t.Log(time.Now().Sub(since1))

	var sc2 = make(chan struct{})
	var w2 = newWork(setIdleTimeout(time.Second), setEndSignalChan(sc2))
	w2.Run(context.Background())
	var since2 = time.Now()
	w2.InputTask() <- Task{
		Message: since2,
		Handler: func(msg interface{}) error {
			ti, _ := msg.(time.Time)
			time.Sleep(time.Second * 2)
			t.Log(time.Now().Sub(ti))
			return nil
		},
	}
	<-sc2
	t.Log(time.Now().Sub(since2))
}

type testLogger struct {
}

func (tl *testLogger) Println(v ...interface{}) {
	fmt.Println(v...)
}

func TestSetLogger(t *testing.T) {
	var esc = make(chan struct{})
	var w = newWork(setLogger(&testLogger{}), setEndSignalChan(esc), setIdleTimeout(time.Second))
	w.Run(context.Background())
	w.InputTask() <- Task{
		Message: "test msg",
		Handler: func(msg interface{}) error {
			panic(msg)
		},
	}
	<-esc
}

func TestSetTaskChanel(t *testing.T) {
	var esc = make(chan struct{})
	var tc = make(chan Task)
	var w = newWork(setTaskChanel(tc), setEndSignalChan(esc), setIdleTimeout(time.Second))
	w.Run(context.Background())
	tc <- Task{
		Message: "test msg",
		Handler: func(msg interface{}) error {
			t.Log(msg)
			return nil
		},
	}
	<-esc
}

func TestSetTaskChanelClose(t *testing.T) {
	var esc = make(chan struct{})
	var tc = make(chan Task)
	var w = newWork(setTaskChanel(tc), setEndSignalChan(esc))
	w.Run(context.Background())
	tc <- Task{
		Message: "test msg",
		Handler: func(msg interface{}) error {
			t.Log(msg)
			return nil
		},
	}
	close(tc)
	<-esc
}

func TestWork_RunWithCTXCancel(t *testing.T) {
	var sc1 = make(chan struct{})
	var ctx1, cancel1 = context.WithCancel(context.Background())
	var w1 = newWork(setEndSignalChan(sc1))
	w1.Run(ctx1)
	w1.InputTask() <- Task{
		Message: "test msg",
		Handler: func(msg interface{}) error {
			t.Log(msg)
			return nil
		},
	}
	cancel1()
	<-sc1

	var sc2 = make(chan struct{})
	var ctx2, cancel2 = context.WithCancel(context.Background())
	var w2 = newWork(setEndSignalChan(sc2))
	w2.Run(ctx2)
	w2.InputTask() <- Task{
		Message: "test msg2",
		Handler: func(msg interface{}) error {
			time.Sleep(time.Second * 2)
			t.Log(msg)
			return nil
		},
	}
	cancel2()
	<-sc2
}
