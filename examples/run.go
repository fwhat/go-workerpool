package main

import (
	"github.com/Dowte/go-workerpool"
	"log"
	"math/rand"
	"runtime"
	"time"
)

type Job struct {
	Id int
}

func (job Job) Run(worker go_workerpool.Worker) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Second * time.Duration(rand.Intn(10)))
	log.Printf("job %d: 执行完成", job.Id)
}

func addQueue(dispatcher *go_workerpool.Dispatcher) {
	for i := 0; i < 20; i++ {
		// 新建一个任务
		job := Job{Id: i}
		// 任务放入任务队列channel
		success := dispatcher.TryEnqueue(job)
		if success {
			log.Printf("job %d: 加入JobQueue队列 成功", job.Id)
		} else {
			log.Printf("job %d: 加入JobQueue队列 失败", job.Id)
		}
	}
}

func main() {
	MaxWorker := 5
	MaxPending := 10
	dispatcher := go_workerpool.NewDispatcher(MaxWorker, MaxPending)
	dispatcher.Run()
	log.Printf("初始化队列数: %d", len(dispatcher.Workers))
	log.Printf("当前协程数: %d", runtime.NumGoroutine())

	go func(dispatcher *go_workerpool.Dispatcher) {
		for {
			log.Printf("空闲的worker数: %d, 当前协程数: %d, 等待处理job数: %d", len(dispatcher.FreeWorkers), runtime.NumGoroutine(), len(dispatcher.PendingJobs))
			time.Sleep(2 * time.Second)
		}
	}(dispatcher)
	time.Sleep(1 * time.Second)
	go addQueue(dispatcher)
	time.Sleep(50 * time.Second)
	go addQueue(dispatcher)
	time.Sleep(1000 * time.Second)

	// 最后协程数: main进程1, 定时统计1, 主调度1协程, worker ${MaxWorker} = 8;
	// 空闲的worker数: ${MaxWorker};
}
