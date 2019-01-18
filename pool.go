package gopool

import "sync"

type Pool struct {
	ch chan struct{} //控制协程池大小，
	wg *sync.WaitGroup
}

// 实例化
// poolSize 协程池大小
// wgSize WaitGroup大小，为0时不等待
func New(poolSize int) *Pool {
	p := &Pool{
		ch: make(chan struct{}, poolSize),
		wg: &sync.WaitGroup{},
	}

	return p
}

// 提交任务
func (p *Pool) Submit(task func()) {
	p.ch <- struct{}{}
	p.wg.Add(1)
	go func() {
		defer func() {
			p.wg.Done()
			<-p.ch
		}()
		task()
	}()
}

// 等待WaitGroup执行完毕
func (p *Pool) Wait() {
	p.wg.Wait()
}
