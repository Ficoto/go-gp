package gp

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	p := New()
	p.Run()
	err := p.Go(func() error {
		time.Sleep(time.Second * 2)
		t.Log("test Go")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = p.GoWithMessage(Task{
		Message: "test",
		Handler: func(msg interface{}) error {
			time.Sleep(time.Second * 2)
			t.Log(msg)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 3)
}

func TestNewWithClose(t *testing.T) {
	p := New()
	p.Run()
	err := p.Go(func() error {
		time.Sleep(time.Second * 2)
		t.Log("test Go")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = p.GoWithMessage(Task{
		Message: "test",
		Handler: func(msg interface{}) error {
			time.Sleep(time.Second * 2)
			t.Log(msg)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	p.Close()
}

func TestNewWithSize(t *testing.T) {
	p1 := New(SetMaxPoolSize(10))
	p1.Run()
	t.Log(p1.Size())
	for i := 0; i != 20; i++ {
		err := p1.Go(func() error {
			time.Sleep(time.Second)
			t.Log("test")
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(p1.Size())
	}
	p1.Close()
	t.Log(p1.Size())
}

func TestPool_Size(t *testing.T) {
	p := New(SetMaxPoolSize(10), SetIdleTimeout(time.Second*2))
	p.Run()
	for i := 0; i != 10; i++ {
		err := p.GoWithMessage(Task{
			Message: i,
			Handler: func(msg interface{}) error {
				s, ok := msg.(int)
				if !ok {
					return errors.New("msg type not right")
				}
				time.Sleep(time.Second * time.Duration(s))
				return nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("pool size:", p.Size())
	}
	time.Sleep(time.Second)
	for i := 0; i != 10; i++ {
		time.Sleep(time.Second)
		t.Log("pool size:", p.Size())
	}
	p.Close()
}

func TestPool_Callback(t *testing.T) {
	p := New(SetCallback(func(task Task, err error) {
		t.Logf("task msg:%v,err:%v", task.Message, err)
	}))
	p.Run()
	for i := 0; i != 10; i++ {
		err := p.GoWithMessage(Task{
			Message: i,
			Handler: func(msg interface{}) error {
				if n, ok := msg.(int); ok && n%2 == 0 {
					return nil
				}
				return errors.New(fmt.Sprintf("error num:%v", msg))
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	p.Close()
}

func TestPool_ErrorCallback(t *testing.T) {
	p := New(SetErrorCallback(func(task Task, err error) {
		t.Logf("task msg:%v,err:%v", task.Message, err)
	}))
	p.Run()
	for i := 0; i != 10; i++ {
		err := p.GoWithMessage(Task{
			Message: i,
			Handler: func(msg interface{}) error {
				if n, ok := msg.(int); ok && n%2 == 0 {
					return nil
				}
				return errors.New(fmt.Sprintf("error num:%v", msg))
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	p.Close()
}

type logger struct {
}

func (l *logger) Println(v ...interface{}) {
	fmt.Println(v...)
}

func TestPoll_AsyncGo(t *testing.T) {
	p := New(SetMaxPoolSize(1), SetLogger(&logger{}))
	p.Run()
	var handler = func() error {
		time.Sleep(time.Second)
		return nil
	}
	now := time.Now()
	p.Go(handler)
	p.Go(handler)
	t.Log(p.Size())
	go func() {
		p.Go(handler)
		t.Logf("this is sync,since %f", time.Since(now).Seconds())
	}()
	go func() {
		p.AsyncGo(handler)
		t.Logf("this is async,since %f", time.Since(now).Seconds())
	}()
	t.Log(p.Size())
	time.Sleep(time.Second * 3)
	p.Close()
}
