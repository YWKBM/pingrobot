package pingrobot

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Worker struct {
	id       int
	taskChan chan *WebServiceInfo
	client   *http.Client
	quit     chan bool
}

func newWorker(id int, tasks chan *WebServiceInfo, timeout time.Duration) *Worker {
	return &Worker{
		id:       id,
		taskChan: tasks,
		client: &http.Client{
			Timeout: timeout,
		},
		quit: make(chan bool),
	}
}

func (w *Worker) StartBackground(wg *sync.WaitGroup, results chan Result) {
	for {
		select {
		case task := <-w.taskChan:
			results <- w.process(w.id, *task, wg)
		case <-w.quit:
			return
		}
	}
}

// TODO: Compile link with port
func (w *Worker) process(workerId int, task WebServiceInfo, wg *sync.WaitGroup) Result {
	wg.Add(1)
	defer wg.Done()

	var url string
	if task.Port != 0 {
		url += task.Link + ":" + fmt.Sprint(task.Port)
	} else {
		url = task.Link
	}

	res := Result{
		ID:        task.ID,
		UserEmail: task.UserEmail,
		URL:       url,
		Alarm:     task.Alarm,
	}

	now := time.Now()

	resp, err := w.client.Get(task.Link)
	if err != nil {
		res.Error = err
		return res
	}

	res.StatusCode = resp.StatusCode
	res.ResponseTime = time.Since(now)

	return res
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
