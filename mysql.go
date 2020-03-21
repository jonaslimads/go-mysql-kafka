package kafka

import (
	"fmt"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"log"
)

type MySqlConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	Tables   []string
}

func (config *MySqlConfig) GetCanalConfig() *canal.Config {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = config.getAddress()
	cfg.User = config.User
	cfg.Password = config.Password
	cfg.Dump.ExecutionPath = ""
	cfg.Dump.TableDB = config.Database
	cfg.Dump.Tables = []string{"employees", "titles"}
	return cfg
}

func (config *MySqlConfig) GetBinlogPosition() mysql.Position {
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
