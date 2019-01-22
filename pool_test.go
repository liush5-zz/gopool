package gopool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	_   = 1 << (10 * iota)
	KiB  // 1024
	MiB  // 1048576
)

const (
	poolSize = 10000
	n        = 1000000
)

var curMem uint64

func demoFunc() {
	n := 10
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func TestPool(t *testing.T) {
	pool := New(poolSize)
	defer pool.Exit()

	start := time.Now()

	t.Log("start at", start.String())

	for i := 0; i < n; i++ {
		pool.Submit(demoFunc)
	}

	pool.Wait()

	t.Log("finish at", time.Now().String())
	t.Logf("pool, capacity:%d", pool.poolSize)
	t.Logf("pool, running workers number:%d", len(pool.workerPool))
	t.Logf("pool, free workers number:%d", len(pool.poolChan))
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem := mem.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", curMem)

}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", curMem)
}
