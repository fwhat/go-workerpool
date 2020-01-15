package main

import (
	"github.com/Dowte/go-workerpool"
	"log"
	"math/rand"
	"runtime"
	"time"
)

type Job struct {
	Id      int
	Process int
}

func (job *Job) Run(worker go_workerpool.Worker) {
	rand.Seed(time.Now().UnixNano())
	total := rand.Intn(10) + 10
	for i := 0; i < rand.Intn(10); i++ {
		time.Sleep(time.Second * time.Duration(i))
		job.Process = i * 100 / total
	}
	job.Process = 100
}

func addQueue(dispatcher *go_workerpool.Dispatcher) (jobs []*Job) {
	for i := 0; i < 20; i++ {
		// 新建一个任务
		job := Job{Id: i}
		// 任务放入任务队列channel
		if dispatcher.TryEnqueue(&job) {
			jobs = append(jobs, &job)
			log.Printf("job %d: 加入JobQueue队列 成功", job.Id)
		} else {
			log.Printf("job %d: 加入JobQueue队列 失败", job.Id)
		}
	}

	return jobs
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
	jobs := addQueue(dispatcher)

	go func(jobs []*Job) {
		for {
			for _, job := range jobs {
				log.Printf("job: %d, process %d", job.Id, job.Process)
			}
			time.Sleep(2 * time.Second)
		}
	}(jobs)

	time.Sleep(1000 * time.Second)

	// 最后协程数: main进程1, 定时统计1, 主调度1协程, worker ${MaxWorker} = 8;
	// 空闲的worker数: ${MaxWorker};
}
