package pingrobot

import (
	"database/sql"
	"fmt"
	"log"
	"pingrobot/email"
	"sort"
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
	Alarm     string `json:"alarm"`
}

type Result struct {
	ID           int
	UserEmail    string
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	Alarm        string
}

type Pool struct {
	db            *sql.DB
	smtp          *email.SMTPSender
	workersCount  int
	tasks         chan *WebServiceInfo
	results       chan Result
	workers       []*Worker
	webServices   []*WebServiceInfo
	mu            sync.Mutex
	resultsToSend []Result
	wg            sync.WaitGroup
}

func NewPool(db *sql.DB, workersCount int, tasks chan *WebServiceInfo, results chan Result, smtp *email.SMTPSender) *Pool {
	return &Pool{
		db:            db,
		smtp:          smtp,
		workersCount:  workersCount,
		tasks:         tasks,
		results:       results,
		resultsToSend: make([]Result, 0, 128),
		wg:            sync.WaitGroup{},
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

		err := rows.Scan(&webService.ID, &webService.UserId, &webService.UserEmail, &webService.Name, &webService.Link, &webService.Port, &webService.Status, &webService.Alarm)
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
		//send Email
		go func() {
			if len(p.resultsToSend) != 0 {
				sort.Slice(p.resultsToSend, func(i, j int) bool {
					return p.resultsToSend[i].UserEmail < p.resultsToSend[j].UserEmail
				})

				p.processResutlToSend(p.resultsToSend)
				p.resultsToSend = make([]Result, 0, 128)
			}
		}()
	}
}

func (p *Pool) processResults() {
	for {
		select {
		case result, ok := <-p.results:
			if !ok {
				log.Fatal()
			}
			if result.Error != nil {
				if result.Alarm == "NOT_ALARMED" {
					p.resultsToSend = append(p.resultsToSend, result)
					//TODO: Set status 'ALARMED' after alarmed processed
					p.db.Query("UPDATE web_services SET alarm = 'ALARMED' WHERE ID = $1", result.ID)
				}
				p.db.Query("UPDATE web_services SET status = 'ERROR' WHERE ID = $1", result.ID)
			} else {
				p.db.Query("UPDATE web_services SET status = 'SUCCESS' WHERE ID = $1", result.ID)
				if result.Alarm == "ALARMED" {
					p.db.Query("UPDATE web_services SET alarm = 'NOT_ALARMED' WHERE ID = $1", result.ID)
				}
			}
			fmt.Println(result)
		default:
			fmt.Println("empty result chanel")
			time.Sleep(time.Millisecond * 10)
		}
	}

}

func (p *Pool) processResutlToSend(results []Result) {
	pivot := 0
	for i := 0; i < len(results); i++ {
		if results[i].UserEmail != results[pivot].UserEmail {
			sendEmailInput := email.NewSendEmailInput(results[pivot].UserEmail, "ERROR")
			sendEmailInput.GenerateBody("templates/error_web_service_email.html", results[pivot:i])
			p.smtp.SendMessage(*sendEmailInput)
			pivot = i
		} else if i == len(results)-1 {
			sendEmailInput := email.NewSendEmailInput(results[pivot].UserEmail, "ERROR")
			sendEmailInput.GenerateBody("templates/error_web_service_email.html", results[pivot:i+1])
			p.smtp.SendMessage(*sendEmailInput)
		}
	}
}

func (p *Pool) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
}
