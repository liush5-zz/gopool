package gopool

import (
	"testing"
	"time"
)

func TestPool_Submit(t *testing.T) {
	pool := New(400)
	defer pool.Exit()

	start := time.Now()

	t.Log("start at", start.String())

	taskNum := 500000
	for i := 0; i < taskNum; i++ {
		pool.Submit(func() {
			time.Sleep(time.Millisecond)
		})
	}

	pool.Wait()
	if time.Since(start) < time.Second {
		t.Error("协程池未阻塞")
		return
	}
	t.Log("finish at", time.Now().String())
}
