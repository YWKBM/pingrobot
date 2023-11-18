package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type ConnectionInfo struct {
	Host     string
	Port     string
	Username string
	DBName   string
	SSLMode  string
	Password string
}

func NewPostgresConnection(info ConnectionInfo) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s", info.Host, info.Port, info.Username, info.Password, info.DBName, info.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	db.SetMaxIdleConns(0)
	db.SetConnMaxIdleTime(1)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
