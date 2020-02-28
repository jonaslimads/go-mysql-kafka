package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
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

func (mysqlConfig *MySqlConfig) toBinlogSyncerConfig() replication.BinlogSyncerConfig {
	return replication.BinlogSyncerConfig{
		ServerID: 1,
		Flavor:   "mysql",
		Host:     mysqlConfig.Host,
		Port:     mysqlConfig.Port,
		User:     mysqlConfig.User,
		Password: mysqlConfig.Password,
	}
}

func (mysqlConfig *MySqlConfig) getAddress() string {
	return fmt.Sprintf("%s:%d", mysqlConfig.Host, mysqlConfig.Port)
}

func main() {
	mysqlConfig := NewMySqlConfigFromEnv()

	syncer := replication.NewBinlogSyncer(mysqlConfig.toBinlogSyncerConfig())
	streamer, _ := syncer.StartSync(getBinlogPosition(mysqlConfig))

	for {
		ev, _ := streamer.GetEvent(context.Background())
		log.Println(ev.Header.EventType)
		ev.Dump(os.Stdout)
	}
}

func getBinlogPosition(mysqlConfig *MySqlConfig) mysql.Position {
	conn, err := client.Connect(
		mysqlConfig.getAddress(),
		mysqlConfig.User,
		mysqlConfig.Password,
		mysqlConfig.Database)
	if err != nil {
		log.Panic(err)
	}

	row, _ := conn.Execute("SHOW MASTER STATUS")
	fileName, _ := row.GetStringByName(0, "File")
	position, _ := row.GetIntByName(0, "Position")

	return mysql.Position{Name: fileName, Pos: uint32(position)}
}
