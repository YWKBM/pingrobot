package pingrobot

import (
	"database/sql"
)

func Run(db *sql.DB) {
	go func() {
		results := make(chan Result)
		tasks := make(chan *WebServiceInfo)

		pool := NewPool(db, 5, tasks, results)

		go pool.RunBackground()
		go pool.generateTasks()
		go pool.processResults()
	}()
}
