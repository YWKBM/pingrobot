package workerpool

import (
	"net/http"
	"time"
)

type worker struct {
	client *http.Client
}

func newWorker(timeout time.Duration) *worker {
	return &worker{
		&http.Client{
			Timeout: timeout,
		},
	}
}

func (w worker) process(j Job) Result {
	res := Result{URL: j.URL}

	now := time.Now()

	resp, err := w.client.Get(j.URL)
	if err != nil {
		res.Error = err

		return res
	}

	res.StatusCode = resp.StatusCode
	res.ResponseTime = time.Since(now)

	return res
}
