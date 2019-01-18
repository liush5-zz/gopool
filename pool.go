package gopool

import (
	"sync"
	"time"
)

type Task func()

type Pool struct {
	poolChan      chan *worker //pool的工作池
	taskRecvQueue chan Task    //接收任务队列
	stop          chan int
	wg            *sync.WaitGroup
	poolSize      int
	workerPool    []*worker
}

type worker struct {
	pool      *Pool     //pool的工作池
	taskQueue chan Task //工作协程的任务队列
	stop      chan int  //结束信号
}

func newWorker(p *Pool) *worker {
	return &worker{
		pool:      p,
		taskQueue: make(chan Task),
		stop:      make(chan int),
	}
}

func (w *worker) run() {
	for {
		w.pool.poolChan <- w //将空闲的worker 放回池中

		select {
		case task := <-w.taskQueue:
			task()

			//任务结束，worker 空闲
		case <-w.stop:
			return
		}

	}
}

func (w *worker) submit(task Task) {
	w.taskQueue <- task
	return
}

func (w *worker) close() {
	w.stop <- 1
}

// 实例化
// poolSize 协程池大小
// wgSize WaitGroup大小，为0时不等待
func New(poolSize int) *Pool {
	p := &Pool{
		poolChan:      make(chan *worker, poolSize), //存放空闲的工作协程
		taskRecvQueue: make(chan Task, poolSize),
		stop:          make(chan int),
		wg:            &sync.WaitGroup{}, //等待任务结束
		poolSize:      poolSize,
		workerPool:    []*worker{}, //存放工作协程
	}
	go p.dispatch()

	////新建workers
	//for i := 0; i < poolSize; i++ {
	//	w := newWorker(p)
	//	go w.run()
	//}
	go p.manage()
	
	return p
}

// 提交任务
func (p *Pool) Submit(task Task) {
	//p.ch <- struct{}{} //若pool满了，阻塞直到有空闲，开启新的worker 协程

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

			idleWorker := p.getWorker() //从工作池中取出一个空闲的worker，处理任务
			idleWorker.submit(task)

			//go func() {
			//	defer func() {
			//		//<-p.ch //结束工作线程
			//	}()
			//	task()
			//}()


		case <-p.stop: //结束信号
			//for i := 0; i < cap(p.workerPool); i++ { //结束工作池中所有的工作协程
			//
			//	worker := <-p.workerPool
			//	close(worker.stop)
			//}

			return
		}

	}
}

func (p *Pool) manage() {
	//init
	if len(p.workerPool) == 0 {

		for i := 0; i < p.poolSize; i++ {
			w := newWorker(p)
			go w.run()
		}
	}

	tick := time.NewTicker(time.Second)
	//resize
	for {
		select {
		case <-tick.C:
			if len(p.workerPool) == p.poolSize {

			} else if len(p.workerPool) > p.poolSize {
				// contract
				idleWork := p.getWorker()
				idleWork.close()

			} else {
				// expand

				w := newWorker(p)
				go w.run()

			}
		case <-p.stop:
			return
		}
	}

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
