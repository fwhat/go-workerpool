package tests

import (
	"context"
	go_workerpool "github.com/Dowte/go-workerpool"
	"log"
	"testing"
	"time"
)

type JobHandleContext struct {
	Id      int
	Process int
	Stop    context.CancelFunc
}

func (job *JobHandleContext) Run(worker go_workerpool.Worker) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	job.Stop = cancelFunc
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			log.Printf("job %d: stopping", job.Id)
			return
		default:
			time.Sleep(time.Second * time.Duration(1))
			job.Process = i * 100 / 10
			log.Printf("job %d: running %d", job.Id, job.Process)
		}
	}
	job.Process = 100
}

func TestStopJob(t *testing.T) {
	MaxWorker := 1
	dispatcher := go_workerpool.NewDispatcher(MaxWorker, 0, nil)
	dispatcher.Run()
	job := &JobHandleContext{Id: 1}
	if !dispatcher.TryEnqueue(job) {
		t.Errorf("job %d: 加入JobQueue队列 失败", job.Id)
	}
	time.Sleep(time.Second)
	if job.Stop == nil {
		t.Errorf("job %d: init stop error", job.Id)
	}
	job.Stop()
	dispatcher.Wait()
}
