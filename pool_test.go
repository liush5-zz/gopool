package gopool

import (
	"runtime"
	"testing"
	"time"
)
const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
)
var curMem uint64

func demoFunc() {
	n := 10
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func TestPool_Submit(t *testing.T) {
	pool := New(400)
	defer time.Sleep(time.Second*5)
	defer pool.Exit()

	start := time.Now()

	t.Log("start at", start.String())

	taskNum := 100000
	for i := 0; i < taskNum; i++ {
		pool.Submit(demoFunc)
	}

	pool.Wait()
	if time.Since(start) < time.Second {
		t.Error("协程池未阻塞")
		return
	}
	t.Log("finish at", time.Now().String())
	t.Logf("pool, capacity:%d", pool.poolSize)
	t.Logf("pool, running workers number:%d", len(pool.workerPool))
	t.Logf("pool, free workers number:%d", len(pool.poolChan))
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem := mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)

}

