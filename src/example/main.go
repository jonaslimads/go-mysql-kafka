package main

import (
	mysqlcdc "github.com/jonaslimads/mysql-cdc"
	"github.com/jonaslimads/mysql-cdc/kafka"
	"log"
	"os"
	"strconv"
)

func main() {
	eventListener := mysqlcdc.NewEventListener(newMySqlConfigFromEnv(), newKafkaStreamer())

	eventListener.HandleChannelFunc("employees", []string{"employees", "titles"}, func(
		streamer mysqlcdc.EventStreamer, event *mysqlcdc.Event) {
		if event.IsUpdate() {
			if err := streamer.Stream("employees", event); err != nil {
				log.Println(err)
			}
		}
	})

	eventListener.Listen()
}

func newMySqlConfigFromEnv() mysqlcdc.MySqlConfig {
	mysqlPort, _ := strconv.Atoi(os.Getenv("MYSQL_PORT"))

	return mysqlcdc.MySqlConfig{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     uint16(mysqlPort),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}
}

func newKafkaStreamer() mysqlcdc.EventStreamer {
	kafkaPort, _ := strconv.Atoi(os.Getenv("KAFKA_PORT"))
	return kafka.NewStreamer(os.Getenv("KAFKA_HOST"), uint16(kafkaPort))
}
