package main

import (
	"fmt"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"log"
	"os"
	"strconv"
)

type MySqlConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
}

func NewMySqlConfigFromEnv() *MySqlConfig {
	mysqlPort, _ := strconv.Atoi(os.Getenv("MYSQL_PORT"))

	return &MySqlConfig{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     uint16(mysqlPort),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}
}

func (config *MySqlConfig) toBinlogSyncerConfig() replication.BinlogSyncerConfig {
	return replication.BinlogSyncerConfig{
		ServerID: 1,
		Flavor:   "mysql",
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
	}
}

func (config *MySqlConfig) getAddress() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func getBinlogPosition(config *MySqlConfig) mysql.Position {
	conn, err := client.Connect(config.getAddress(), config.User, config.Password, config.Database)
	if err != nil {
		log.Panic(err)
	}

	row, _ := conn.Execute("SHOW MASTER STATUS")
	fileName, _ := row.GetStringByName(0, "File")
	position, _ := row.GetIntByName(0, "Position")

	return mysql.Position{Name: fileName, Pos: uint32(position)}
}
