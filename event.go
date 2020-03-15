package mysqlcdc

import (
	"bytes"
	"context"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/replication"
	"regexp"
	"strconv"
)

type EventListener struct {
	config   MySqlConfig
	streamer EventStreamer
	channels []*Channel
	tables   map[uint64]string
}

type EventStreamer interface {
	Stream(topic string, event *Event) error
}

type Event struct {
	BinlogEvent *replication.BinlogEvent
	Action      EventAction
	Table       *Table
}

type Table struct {
	id   uint64
	name string
}

type EventAction string

var InsertAction EventAction = canal.InsertAction

var DeleteAction EventAction = canal.DeleteAction

var UpdateAction EventAction = canal.UpdateAction

func NewEventListener(config MySqlConfig, streamer EventStreamer) *EventListener {
	return &EventListener{
		config:   config,
		channels: make([]*Channel, 0),
		tables:   make(map[uint64]string),
		streamer: streamer,
	}
}

func (eventListener *EventListener) HandleChannel(topic string, tables []string, streamHandler ChannelHandler) *Channel {
	return eventListener.NewChannel().Topic(topic).Tables(tables).ChannelHandler(streamHandler)
}

func (eventListener *EventListener) HandleChannelFunc(topic string, tables []string,
	f func(EventStreamer, *Event)) *Channel {
	return eventListener.NewChannel().Topic(topic).Tables(tables).ChannelHandlerFunc(f)
}

func (eventListener *EventListener) NewChannel() *Channel {
	channel := &Channel{}
	eventListener.channels = append(eventListener.channels, channel)
	return channel
}

func (eventListener *EventListener) Listen() {
	syncer := replication.NewBinlogSyncer(eventListener.config.toBinlogSyncerConfig())
	listener, _ := syncer.StartSync(eventListener.config.getBinlogPosition())

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
	event := &Event{BinlogEvent: bingLogEvent, Action: action, Table: table}
	for _, channel := range eventListener.MatchChannels(event) {
		go channel.handler.HandleEvent(eventListener.streamer, event)
	}
}

func (eventListener *EventListener) MatchChannels(event *Event) []*Channel {
	matchedChannels := make([]*Channel, 0)
	for _, channel := range eventListener.channels {
		if channel.Match(event) {
			matchedChannels = append(matchedChannels, channel)
		}
	}
	return matchedChannels
}

func (event *Event) IsInsert() bool {
	return event.Action == InsertAction
}

func (event *Event) IsUpdate() bool {
	return event.Action == UpdateAction
}

func (event *Event) IsDelete() bool {
	return event.Action == DeleteAction
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
