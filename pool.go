package gopool

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultSize = 10
)

type Task func()

type Pool struct {
	poolChan      chan *worker //pool的工作池
	taskRecvQueue chan Task    //接收任务队列
	stop          chan int
	wg            *sync.WaitGroup
	poolSize      int64
	mux           sync.Mutex
	workerPool    map[*worker]bool
}

type worker struct {
	pool      *Pool     //pool的工作池
	taskQueue chan Task //工作协程的任务队列
	stop      chan int  //结束信号
}

func newWorker(p *Pool) *worker {
	w := &worker{
		pool:      p,
		taskQueue: make(chan Task, 1),
		stop:      make(chan int),
	}

	//在工作池中注册
	p.mux.Lock()
	p.workerPool[w] = true
	p.mux.Unlock()
	return w
}

func (w *worker) run() {
	for {
		w.pool.poolChan <- w //将空闲的worker 放回池中

		select {
		case task := <-w.taskQueue:
			task()
			//任务结束，worker 空闲

		case <-w.stop:

			//在工作池中注销
			w.pool.mux.Lock()
			delete(w.pool.workerPool, w)
			w.pool.mux.Unlock()

			return
		}

	}
}

func (w *worker) submit(task Task) {
	w.taskQueue <- task
	return
}

func (w *worker) close() {
	close(w.stop)
}

// 实例化
// poolSize 协程池大小
func New(poolSize int64) *Pool {
	if poolSize <= 0 {
		poolSize = DefaultSize
	}
	p := &Pool{
		poolChan:      make(chan *worker, poolSize), //存放空闲的工作协程
		taskRecvQueue: make(chan Task, poolSize),
		stop:          make(chan int),
		wg:            &sync.WaitGroup{}, //等待任务结束
		poolSize:      poolSize,
		mux:           sync.Mutex{},
		workerPool:    make(map[*worker]bool), //存放工作协程
	}

	go p.dispatch()
	go p.manage()

	return p
}

// 提交任务
func (p *Pool) Submit(task Task) {

	p.wg.Add(1)
	taskWrap := func() {
		task()
		p.wg.Done()
	}

	p.taskRecvQueue <- taskWrap

}

//任务的分发
func (p *Pool) dispatch() {

	for {
		select {
		case task := <-p.taskRecvQueue: //取出任务

			//从工作池中取出一个空闲的worker，处理任务
			idleWorker := p.getWorker()
			idleWorker.submit(task)

		case <-p.stop: //结束信号
			return
		}

	}
}

func (p *Pool) manage() {
	//init
	if len(p.workerPool) == 0 {

		for i := int64(0); i < p.poolSize; i++ {
			w := newWorker(p)
			go w.run()
		}
	}

	//resize
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			currentSize := int64(len(p.workerPool))
			if currentSize == p.poolSize {

			} else if currentSize > p.poolSize {
				// reduce
				idleWork := p.getWorker()
				idleWork.close()

			} else {
				// expand
				w := newWorker(p)
				go w.run()

			}
		case <-p.stop:
			p.mux.Lock()
			for w := range p.workerPool { //结束工作池中所有的工作协程
				w.close()
			}
			p.mux.Unlock()

			return
		}
	}

	return

}

func (p *Pool) Resize(newSize int64) {
	atomic.StoreInt64(&p.poolSize, newSize)
	return
}

//获取空闲worker
func (p *Pool) getWorker() *worker {
	worker := <-p.poolChan
	return worker
}
func (p *Pool) GetSize() int {
	return len(p.poolChan)
}

// 等待WaitGroup执行完毕
func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Exit() {
	close(p.stop)
}
