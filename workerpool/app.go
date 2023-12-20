package pingrobot

import (
	"database/sql"
	"pingrobot/email"
)

func Run(db *sql.DB, smtpSender *email.SMTPSender) {
	go func() {
		results := make(chan Result)
		tasks := make(chan *WebServiceInfo)

		pool := NewPool(db, 5, tasks, results, smtpSender)

		go pool.RunBackground()
		go pool.generateTasks()
		go pool.processResults()
	}()
}
