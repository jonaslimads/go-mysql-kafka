package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/siddontang/go-mysql/canal"
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

type EventAction string

var InsertAction EventAction = canal.InsertAction

var DeleteAction EventAction = canal.DeleteAction

var UpdateAction EventAction = canal.UpdateAction

type EventStreamer struct {
	mysqlConfig *MySqlConfig
	streams     []*Stream
	tables map[uint64]string
}

type Stream struct {
	topic         string
	path          string
	streamHandler StreamHandler
}

type Event struct {
	event *replication.BinlogEvent
	action EventAction
	table *Table
}

type Table struct {
	id uint64
	name string
}

type StreamHandler interface {
	Stream(*Event)
}

func NewEventStreamer(config *MySqlConfig) *EventStreamer {
	return &EventStreamer{
		mysqlConfig: config,
		streams: make([]*Stream, 0),
		tables: make(map[uint64]string),
	}
}

func (eventStreamer *EventStreamer) HandleStream(topic string, path string, streamHandler StreamHandler) *Stream {
	return eventStreamer.NewStream().Topic(topic).Path(path).StreamHandler(streamHandler)
}

func (eventStreamer *EventStreamer) HandleStreamFunc(topic string, path string,
	f func(*Event)) *Stream {
	return eventStreamer.NewStream().Topic(topic).Path(path).StreamHandlerFunc(f)
}

func (eventStreamer *EventStreamer) NewStream() *Stream {
	stream := &Stream{}
	eventStreamer.streams = append(eventStreamer.streams, stream)
	return stream
}

func (eventStreamer *EventStreamer) run() {
	syncer := replication.NewBinlogSyncer(eventStreamer.mysqlConfig.toBinlogSyncerConfig())
	//syncer
	streamer, _ := syncer.StartSync(getBinlogPosition(eventStreamer.mysqlConfig))

	for {
		event, _ := streamer.GetEvent(context.Background())
		eventStreamer.EmitEvent(event)
	}
}

var tableRegex = regexp.MustCompile(`(?m)(?m)^(TableID|Table):\s*(\w+)`)

func (eventStreamer *EventStreamer) EmitEvent(event *replication.BinlogEvent) {
	stream := eventStreamer.MatchStream(event)
	if stream == nil {
		return
	}

	if event.Header.EventType == replication.TABLE_MAP_EVENT {
		tableID, tableName := getTableIDName(event)
		eventStreamer.tables[tableID] = tableName
		return
	}

	action := getEventAction(event)
	if action == "" {
		return
	}

	tableID, _ := getTableIDName(event)
	table := &Table{id: tableID, name: eventStreamer.tables[tableID]}
	stream.streamHandler.Stream(&Event{event: event, action: action, table: table})
}

func getTableIDName(event *replication.BinlogEvent) (uint64, string) {
	tableID := -1
	tableName := ""

	buf := new(bytes.Buffer)
	event.Dump(buf)

	for _, match := range tableRegex.FindAllStringSubmatch (buf.String(), -1) {
		if match[1] == "TableID" {
			tableID, _ = strconv.Atoi(match[2])
		} else if match[1] == "Table" {
			tableName = match[2]
		}
	}
	return uint64(tableID), tableName
}

func getEventAction(event *replication.BinlogEvent) EventAction {
	switch event.Header.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		return InsertAction
	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		return DeleteAction
	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		return UpdateAction
	}
	return ""
}

func (eventStreamer *EventStreamer) MatchStream(event *replication.BinlogEvent) *Stream {
	for _, stream := range eventStreamer.streams {
		if stream.Match(event) {
			return stream
		}
	}
	return nil
}

func (stream *Stream) Topic(topic string) *Stream {
	stream.topic = topic
	return stream
}

func (stream *Stream) Path(path string) *Stream {
	stream.path = path
	return stream
}

func (stream *Stream) StreamHandler(streamHandler StreamHandler) *Stream {
	stream.streamHandler = streamHandler
	return stream
}

func (stream *Stream) StreamHandlerFunc(f func(*Event)) *Stream {
	return stream.StreamHandler(StreamHandlerFunc(f))
}

func (stream *Stream) Match(event *replication.BinlogEvent) bool {
	return stream.path == "employees"
}

type StreamHandlerFunc func(*Event)

func (streamHandlerFunc StreamHandlerFunc) Stream(event *Event) {
	streamHandlerFunc(event)
}

func main() {
	eventStreamer := NewEventStreamer(NewMySqlConfigFromEnv())
	eventStreamer.HandleStreamFunc("employees", "employees", func(event *Event) {
		if event.action == UpdateAction {
			log.Println("got update event", event.table, event.action)
			event.event.Dump(os.Stdout)
		}
	})
	eventStreamer.run()
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
