package gopool

import "sync"

type Task func()

type Pool struct {
	ch            chan struct{} //控制协程池大小，
	taskRecvQueue chan Task     //接收任务队列
	stop          chan int
	wg            *sync.WaitGroup
}

// 实例化
// poolSize 协程池大小
// wgSize WaitGroup大小，为0时不等待
func New(poolSize int) *Pool {
	p := &Pool{
		ch:            make(chan struct{}, poolSize),
		taskRecvQueue: make(chan Task, poolSize),
		stop:          make(chan int),
		wg:            &sync.WaitGroup{},
	}
	go p.dispatch()
	return p
}

// 提交任务
func (p *Pool) Submit(task Task) {
	p.ch <- struct{}{}

	p.taskRecvQueue <- task

}

//任务的分发
func (p *Pool) dispatch() {

	for {
		select {
		case task := <-p.taskRecvQueue: //取出任务

			//worker := <-p.workerPool //从工作池中取出一个worker，处理任务
			//worker.jobQueue <- job
			p.wg.Add(1)
			go func() {
				defer func() {
					p.wg.Done()
					<-p.ch
				}()
				task()
			}()


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

// 等待WaitGroup执行完毕
func (p *Pool) Wait() {
	p.wg.Wait()
}

/*func (p *Pool) Exit() {
	close(p.stop)
}*/
