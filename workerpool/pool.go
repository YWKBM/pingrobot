package workerpool

import (
	"fmt"
	"log"
	"time"
)

type Job struct {
	URL string
}

type Result struct {
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

type Pool struct {
	worker       *worker
	workersCount int

	jobs   chan Job
	result chan Result

	stopped bool
}

func New(workersCount int, timeout time.Duration, results chan Result) *Pool {
	return &Pool{
		worker:       newWorker(timeout),
		workersCount: workersCount,
		jobs:         make(chan Job),
		result:       results,
	}
}

func (p *Pool) Init() {
	for i := 0; i < p.workersCount; i++ {
		go p.initWorker(i)
	}
}

func (r Result) Info() string {
	if r.Error != nil {
		return fmt.Sprintf("[ERROR] - [%s] - %s", r.URL, r.Error.Error())
	}

	return fmt.Sprintf("[SUCCESS] - [%s] - Status: %d, Response Time: %s", r.URL, r.StatusCode, r.ResponseTime.String())
}
func (p *Pool) Push(j Job) {
	if p.stopped {
		return
	}

	p.jobs <- j
}

func (p *Pool) Stop() {
	p.stopped = true
	close(p.jobs)
}

func (p *Pool) initWorker(id int) {
	for job := range p.jobs {
		time.Sleep(time.Second)
		p.result <- p.worker.process(job)
	}

	log.Printf("[worker ID %d] finished processing", id)
}
