package gp

import (
	"runtime"
	"sync"
	"testing"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
)

const (
	Param    = 100
	Size     = 1000
	TestSize = 10000
	n        = 100000
)

var curMem uint64

func TestGPMemStats(t *testing.T) {
	p := New(SetMaxPoolSize(Size))
	for i := 0; i < n; i++ {
		_ = p.Go(func() {
			demoPoolFunc(Param)
		})
	}
	t.Logf("pool size %d", p.Size())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
	p.Close()
}

func TestGoroutineMemStats(t *testing.T) {
	var (
		wg sync.WaitGroup
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoPoolFunc(Param)
			wg.Done()
		}()
	}
	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}
