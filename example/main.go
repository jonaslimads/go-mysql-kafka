package main

import (
	kafka "github.com/jonaslimads/go-mysql-kafka"
	"github.com/siddontang/go-mysql/canal"
	"log"
	"os"
	"strconv"
)

func main() {
	eventHandler := kafka.NewEventHandler(newMySqlConfigFromEnv(), newKafkaStreamer())

	eventHandler.HandleChannelFunc("employees", []string{"employees", "titles", "salaries"}, func(
		streamer kafka.EventStreamer, event *canal.RowsEvent) {
		if err := streamer.Stream("employees", event); err != nil {
			log.Println(err)
		}
	})

	eventHandler.Run()
}

func newMySqlConfigFromEnv() kafka.MySqlConfig {
	mysqlPort, _ := strconv.Atoi(os.Getenv("MYSQL_PORT"))

	return kafka.MySqlConfig{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     uint16(mysqlPort),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}
}

func newKafkaStreamer() kafka.EventStreamer {
	kafkaPort, _ := strconv.Atoi(os.Getenv("KAFKA_PORT"))
	return kafka.NewStreamer(os.Getenv("KAFKA_HOST"), uint16(kafkaPort))
}
