package pingrobot

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

type WebServiceInfo struct {
	ID        int    `json:"id"`
	UserId    int    `json:"user_id"`
	UserEmail string `json:"user_email"`
	Name      string `json:"name"`
	Link      string `json:"link"`
	Port      int    `json:"port"`
	Status    string `json:"status"`
}

type Result struct {
	ID           int
	UserEmail    string
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

type Pool struct {
	db           *sql.DB
	workersCount int
	tasks        chan *WebServiceInfo
	results      chan Result
	workers      []*Worker
	webServices  []*WebServiceInfo
	wg           sync.WaitGroup
}

func NewPool(db *sql.DB, workersCount int, tasks chan *WebServiceInfo, results chan Result) *Pool {
	return &Pool{
		db:           db,
		workersCount: workersCount,
		tasks:        tasks,
		results:      results,
		wg:           sync.WaitGroup{},
	}
}

func (p *Pool) RunBackground() {
	for i := 1; i <= p.workersCount; i++ {
		worker := newWorker(i, p.tasks, 5*time.Second)
		p.workers = append(p.workers, worker)
		go worker.StartBackground(&p.wg, p.results)
	}
}

func (p *Pool) getAllWebServiceInfo() {
	rows, err := p.db.Query("SELECT * FROM web_services")
	if err != nil {
		log.Fatal(err)
	}

	p.webServices = make([]*WebServiceInfo, 128)
	for rows.Next() {
		var webService WebServiceInfo

		err := rows.Scan(&webService.ID, &webService.UserId, &webService.UserEmail, &webService.Name, &webService.Link, &webService.Port, &webService.Status)
		p.webServices = append(p.webServices, &webService)
		if err != nil {
			fmt.Println(err)
		}
	}

	rows.Close()
	return
}

func (p *Pool) generateTasks() {
	for {
		p.getAllWebServiceInfo()
		for i, webService := range p.webServices {
			if webService != nil {
				p.tasks <- webService
				p.webServices[i] = nil
			}
		}
		time.Sleep(60 * time.Second)
	}
}

func (p *Pool) processResults() {
	go func() {
		for {
			result := <-p.results
			fmt.Println(result)
			if result.Error != nil {
				//TODO: send email
				p.db.Query("UPDATE web_services SET status = 'ERROR' WHERE ID = $1", result.ID)
				continue
			}
			p.db.Query("UPDATE web_services SET status = 'SUCCESS' WHERE ID = $1", result.ID)
		}
	}()
}

func (p *Pool) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
}
