package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	eventListener := NewEventListener(newMySqlConfigFromEnv())

	eventListener.HandleStreamFunc("employees", []string{"employees"}, func(event *Event) {
		if event.action == UpdateAction {
			log.Println("=> LOL", event.table, event.action)
			//event.binlogEvent.Dump(os.Stdout)
		}
	})

	eventListener.HandleStreamFunc("employees", []string{"titles"}, func(event *Event) {
		if event.action == UpdateAction {
			log.Println("==> HUE", event.table, event.action)
			//event.binlogEvent.Dump(os.Stdout)
		}
	})

	eventListener.run()
}

func newMySqlConfigFromEnv() *MySqlConfig {
	mysqlPort, _ := strconv.Atoi(os.Getenv("MYSQL_PORT"))

	return &MySqlConfig{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     uint16(mysqlPort),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}
}