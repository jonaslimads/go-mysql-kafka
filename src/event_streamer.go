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

type EventStreamer struct {
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

func NewEventStreamer(config *MySqlConfig) *EventStreamer {
	return &EventStreamer{
		mysqlConfig: config,
		streams:     make([]*Stream, 0),
		tables:      make(map[uint64]string),
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
	streamer, _ := syncer.StartSync(getBinlogPosition(eventStreamer.mysqlConfig))

	for {
		event, _ := streamer.GetEvent(context.Background())
		go eventStreamer.EmitEvent(event)
	}
}

var tableRegex = regexp.MustCompile(`(?m)(?m)^(TableID|Table):\s*(\w+)`)

func (eventStreamer *EventStreamer) EmitEvent(bingLogEvent *replication.BinlogEvent) {
	if bingLogEvent.Header.EventType == replication.TABLE_MAP_EVENT {
		tableID, tableName := getTableIDName(bingLogEvent)
		eventStreamer.tables[tableID] = tableName
		return
	}

	action := getEventAction(bingLogEvent)
	if action == "" {
		return
	}

	tableID, _ := getTableIDName(bingLogEvent)
	table := &Table{id: tableID, name: eventStreamer.tables[tableID]}

	event := &Event{binlogEvent: bingLogEvent, action: action, table: table}
	for _, stream := range eventStreamer.MatchStreams(event) {
		go stream.streamHandler.Stream(event)
	}
}

func (eventStreamer *EventStreamer) MatchStreams(event *Event) []*Stream {
	matchedStreams := make([]*Stream, 0)
	for _, stream := range eventStreamer.streams {
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