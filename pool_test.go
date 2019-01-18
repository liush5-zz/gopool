package gopool

import (
	"testing"
	"time"
)

func TestPool_Submit(t *testing.T) {
	pool := New(2)
	start := time.Now()

	t.Log("start at", start.String())

	taskNum := 5
	for i := 0; i < taskNum; i++ {
		pool.Submit(func() {
			t.Log("线程池大小", pool.GetSize())
			time.Sleep(time.Second)
		})
	}

	pool.Wait()
	if time.Since(start) < time.Second {
		t.Error("协程池未阻塞")
		return
	}
	t.Log("finish at", time.Now().String())
}
