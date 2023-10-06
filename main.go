package main

import (
	"fmt"
	"os"
	"os/signal"
	"pingrobot/workerpool"
	"syscall"
	"time"
)

func main() {
	var urls = []string{
		"https://workshop.zhashkevych.com/",
		"https://golang-ninja.com/",
		"https://zhashkevych.com/",
		"https://google.com/",
		"https://golang.org/",
	}

	results := make(chan workerpool.Result)

	pool := workerpool.New(4, 2*time.Second, results)
	pool.Init()

	go generateJobs(pool, urls)
	go processResults(results)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	pool.Stop()

}

func generateJobs(wp *workerpool.Pool, urls []string) {
	for {
		for _, url := range urls {
			wp.Push(workerpool.Job{URL: url})
		}
	}
}

func processResults(results chan workerpool.Result) {
	go func() {
		for result := range results {
			fmt.Println(result.Info())
		}
	}()
}
