package go_workerpool

type Job interface {
	Run(worker Worker)
}

// execute process
type Worker struct {
	OnSuccessChannel chan Worker // worker执行完成通知 Channel
	JobChannel       chan Job    // 新job分配 Channel
	quit             chan bool   // 退出信号
	no               int         // worker编号
}

//创建一个新worker
func NewWorker(successChannel chan Worker, no int) Worker {
	return Worker{
		OnSuccessChannel: successChannel,
		JobChannel:       make(chan Job),
		quit:             make(chan bool),
		no:               no,
	}
}

//循环  监听任务和结束信号
func (w Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.JobChannel:
				// 收到新的job任务 开始执行
				job.Run(w)
				w.OnSuccessChannel <- w
			case <-w.quit:
				// 收到退出信号
				return
			}
		}
	}()
}

// 停止信号
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

//调度中心
type Dispatcher struct {
	// 空闲的worker
	FreeWorkers chan Worker
	// 所有worker实例
	Workers []Worker
	// 等待执行的job
	PendingJobs chan Job
}

//创建调度中心
func NewDispatcher(maxWorkers int, maxPendingJobs int) *Dispatcher {
	return &Dispatcher{FreeWorkers: make(chan Worker, maxWorkers), PendingJobs: make(chan Job, maxPendingJobs)}
}

//工作者池的初始化
func (d *Dispatcher) Run() {
	for i := 1; i < cap(d.FreeWorkers)+1; i++ {
		worker := NewWorker(d.FreeWorkers, i)
		worker.Start()

		d.FreeWorkers <- worker
		d.Workers = append(d.Workers, worker)
	}
	go d.dispatch()
}

//调度
func (d *Dispatcher) dispatch() {
	for {
		// 判断是否有空闲worker
		if len(d.FreeWorkers) > 0 {
			select {
			// 得到 pending 中的job 和 空闲的worker， 并通知worker开始执行
			case job := <-d.PendingJobs:
				worker := <-d.FreeWorkers
				worker.JobChannel <- job
			}
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
