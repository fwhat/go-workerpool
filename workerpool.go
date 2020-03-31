package go_workerpool

import (
	"context"
	"log"
	"sync"
)

type Job interface {
	Run(worker Worker)
}

// execute process
type Worker struct {
	onSuccessChan chan<- Worker   // worker执行完成通知 Channel
	onJobFinished func(Job)       // job执行完成通知 Channel
	jobChannel    chan Job        // 新job分配 Channel
	no            int             // worker编号
	wg            *sync.WaitGroup // wait group
}

//创建一个新worker
func NewWorker(successChan chan Worker, no int, onJobFinished func(Job), wg *sync.WaitGroup) Worker {
	return Worker{
		onJobFinished: onJobFinished,
		onSuccessChan: successChan,
		jobChannel:    make(chan Job),
		no:            no,
		wg:            wg,
	}
}

//循环  监听任务和结束信号
func (w Worker) Start(ctx context.Context) {
	for {
		select {
		case job := <-w.jobChannel:
			// 收到新的job任务 开始执行
			job.Run(w)
			if w.onJobFinished != nil {
				w.onJobFinished(job)
			}

			w.onSuccessChan <- w
		case <-ctx.Done():
			log.Println("cancel")
			w.wg.Done()
			// 收到退出信号
			return
		}
	}
}

//调度中心
type Dispatcher struct {
	// on job exit
	OnJobExit func(Job)
	// 空闲的worker
	FreeWorkers chan Worker
	// 所有worker实例
	Workers []Worker
	// 等待执行的job
	PendingJobs chan Job

	wg        *sync.WaitGroup
	ctx       context.Context
	ctxCancel context.CancelFunc
}

//创建调度中心
func NewDispatcher(maxWorkers int, maxPendingJobs int, onJobExit func(Job)) *Dispatcher {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Dispatcher{
		FreeWorkers: make(chan Worker, maxWorkers),
		PendingJobs: make(chan Job, maxPendingJobs+maxWorkers),
		OnJobExit:   onJobExit,
		ctxCancel:   cancelFunc,
		ctx:         ctx,
		wg:          &sync.WaitGroup{},
	}
}

//工作者池的初始化
func (d *Dispatcher) Run() {
	for i := 1; i < cap(d.FreeWorkers)+1; i++ {
		worker := NewWorker(d.FreeWorkers, i, d.OnJobExit, d.wg)
		d.wg.Add(1)
		go worker.Start(d.ctx)

		d.FreeWorkers <- worker
		d.Workers = append(d.Workers, worker)
	}
	go d.dispatch()
}

//调度
func (d *Dispatcher) dispatch() {
	for {
		select {
		case worker := <-d.FreeWorkers:
			// 得到 pending 中的job 和 空闲的worker， 并通知worker开始执行
			select {
			case job := <-d.PendingJobs:
				worker.jobChannel <- job
			case <-d.ctx.Done():
				return
			}

		case <-d.ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) TryEnqueue(job Job) bool {
	select {
	case d.PendingJobs <- job:
		return true
	default:
		return false
	}
}

func (d *Dispatcher) Wait() {
	d.ctxCancel()
	d.wg.Wait()
}
