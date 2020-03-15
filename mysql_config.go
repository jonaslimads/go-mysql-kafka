package mysqlcdc

import (
	"fmt"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"log"
)

type MySqlConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
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

func (config *MySqlConfig) getBinlogPosition() mysql.Position {
	conn, err := client.Connect(config.getAddress(), config.User, config.Password, config.Database)
	if err != nil {
		log.Panic(err)
	}

	row, _ := conn.Execute("SHOW MASTER STATUS")
	fileName, _ := row.GetStringByName(0, "File")
	position, _ := row.GetIntByName(0, "Position")

	return mysql.Position{Name: fileName, Pos: uint32(position)}
}

func (config *MySqlConfig) getAddress() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
