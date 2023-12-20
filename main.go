package main

import (
	"log"
	"os"
	"os/signal"
	"pingrobot/database"
	"pingrobot/email"
	workerpool "pingrobot/workerpool"
	"syscall"

	"github.com/spf13/viper"
)

func main() {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	viper.ReadInConfig()

	connectionInfo := database.ConnectionInfo{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		DBName:   viper.GetString("db.dbname"),
		Username: viper.GetString("db.username"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: viper.GetString("db.password"),
	}

	db, err := database.NewPostgresConnection(connectionInfo)
	if err != nil {
		log.Fatal()
	}
	smtpSender := email.NewSMTPSender("ywkbm1973@gmail.com", "ctjf gmqd kcmc xqac", "smtp.gmail.com", 587)

	workerpool.Run(db, smtpSender)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
}
