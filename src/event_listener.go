package main

import (
	"bytes"
	"context"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/replication"
	"regexp"
	"strconv"
)

type EventAction string

var InsertAction EventAction = canal.InsertAction

var DeleteAction EventAction = canal.DeleteAction

var UpdateAction EventAction = canal.UpdateAction

type EventListener struct {
	mysqlConfig *MySqlConfig
	streams     []*Stream
	tables      map[uint64]string
}

type Event struct {
	binlogEvent *replication.BinlogEvent
	action      EventAction
	table       *Table
}

type Table struct {
	id   uint64
	name string
}

func NewEventListener(config *MySqlConfig) *EventListener {
	return &EventListener{
		mysqlConfig: config,
		streams:     make([]*Stream, 0),
		tables:      make(map[uint64]string),
	}
}

func (eventListener *EventListener) HandleStream(topic string, tables []string, streamHandler StreamHandler) *Stream {
	return eventListener.NewStream().Topic(topic).Tables(tables).StreamHandler(streamHandler)
}

func (eventListener *EventListener) HandleStreamFunc(topic string, tables []string,
	f func(*Event)) *Stream {
	return eventListener.NewStream().Topic(topic).Tables(tables).StreamHandlerFunc(f)
}

func (eventListener *EventListener) NewStream() *Stream {
	stream := &Stream{}
	eventListener.streams = append(eventListener.streams, stream)
	return stream
}

func (eventListener *EventListener) run() {
	syncer := replication.NewBinlogSyncer(eventListener.mysqlConfig.toBinlogSyncerConfig())
	listener, _ := syncer.StartSync(eventListener.mysqlConfig.getBinlogPosition())

	for {
		event, _ := listener.GetEvent(context.Background())
		go eventListener.HandleEvent(event)
	}
}

var tableRegex = regexp.MustCompile(`(?m)(?m)^(TableID|Table):\s*(\w+)`)

func (eventListener *EventListener) HandleEvent(bingLogEvent *replication.BinlogEvent) {
	if bingLogEvent.Header.EventType == replication.TABLE_MAP_EVENT {
		tableID, tableName := getTableIDName(bingLogEvent)
		eventListener.tables[tableID] = tableName
		return
	}

	action := getEventAction(bingLogEvent)
	if action == "" {
		return
	}

	tableID, _ := getTableIDName(bingLogEvent)
	table := &Table{id: tableID, name: eventListener.tables[tableID]}

	event := &Event{binlogEvent: bingLogEvent, action: action, table: table}
	for _, stream := range eventListener.MatchStreams(event) {
		go stream.streamHandler.Stream(event)
	}
}

func (eventListener *EventListener) MatchStreams(event *Event) []*Stream {
	matchedStreams := make([]*Stream, 0)
	for _, stream := range eventListener.streams {
		if stream.Match(event) {
			matchedStreams = append(matchedStreams, stream)
		}
	}
	return matchedStreams
}


func getTableIDName(binlogEvent *replication.BinlogEvent) (uint64, string) {
	tableID := -1
	tableName := ""

	buf := new(bytes.Buffer)
	binlogEvent.Dump(buf)

	for _, match := range tableRegex.FindAllStringSubmatch(buf.String(), -1) {
		if match[1] == "TableID" {
			tableID, _ = strconv.Atoi(match[2])
		} else if match[1] == "Table" {
			tableName = match[2]
		}
	}
	return uint64(tableID), tableName
}

func getEventAction(binlogEvent *replication.BinlogEvent) EventAction {
	switch binlogEvent.Header.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		return InsertAction
	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		return DeleteAction
	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		return UpdateAction
	}
	return ""
}