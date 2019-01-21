package gopool

import (
	"sync"
	"testing"
)

const RunTimes = 1000000
const benchAntsSize = 200000

func BenchmarkGoroutine(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {

	p := New(benchAntsSize)
	defer p.Exit()

	//b.StartTimer()
	for i := 0; i < b.N; i++ {

		for j := 0; j < RunTimes; j++ {
			p.Submit(demoFunc)
		}
		p.Wait()
		//b.Logf("running goroutines: %d", p.Running())
	}
	//b.StopTimer()
}
