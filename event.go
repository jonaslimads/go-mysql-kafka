package kafka

import (
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"log"
)

type EventHandler struct {
	config   MySqlConfig
	streamer EventStreamer
	channels []*Channel
}

type EventStreamer interface {
	Stream(topic string, event *canal.RowsEvent) error
}

func NewEventHandler(config MySqlConfig, streamer EventStreamer) *EventHandler {
	return &EventHandler{
		config:   config,
		streamer: streamer,
		channels: make([]*Channel, 0),
	}
}

func (handler *EventHandler) HandleChannel(topic string, tables []string, streamHandler ChannelHandler) *Channel {
	return handler.NewChannel().Topic(topic).Tables(tables).ChannelHandler(streamHandler)
}

func (handler *EventHandler) HandleChannelFunc(topic string, tables []string,
	f func(EventStreamer, *canal.RowsEvent)) *Channel {
	return handler.NewChannel().Topic(topic).Tables(tables).ChannelHandlerFunc(f)
}

func (handler *EventHandler) NewChannel() *Channel {
	channel := &Channel{}
	handler.channels = append(handler.channels, channel)
	return channel
}

func (handler *EventHandler) OnRow(event *canal.RowsEvent) error {
	for _, channel := range handler.MatchChannels(event) {
		go channel.handler.HandleEvent(handler.streamer, event)
	}
	return nil
}

func (handler *EventHandler) MatchChannels(event *canal.RowsEvent) []*Channel {
	matchedChannels := make([]*Channel, 0)
	for _, channel := range handler.channels {
		if channel.Match(event) {
			matchedChannels = append(matchedChannels, channel)
		}
	}
	return matchedChannels
}

func (handler *EventHandler) String() string {
	return "EventHandler"
}

func (handler *EventHandler) Run() {
	handler.config.Tables = handler.getTables()

	c, err := canal.NewCanal(handler.config.GetCanalConfig())
	if err != nil {
		log.Panic(err)
	}

	c.SetEventHandler(handler)
	err = c.RunFrom(handler.config.GetBinlogPosition())
	if err != nil {
		log.Panic(err)
	}
}

func (handler *EventHandler) getTables() []string {
	var tables []string
	for _, channel := range handler.channels {
		tables = append(tables, channel.tables...)
	}
	return tables
}

// methods to meet EventHandler interface

func (handler *EventHandler) OnRotate(*replication.RotateEvent) error {
	return nil
}

func (handler *EventHandler) OnTableChanged(string, string) error {
	return nil
}

func (handler *EventHandler) OnDDL(mysql.Position, *replication.QueryEvent) error {
	return nil
}

func (handler *EventHandler) OnXID(mysql.Position) error {
	return nil
}

func (handler *EventHandler) OnGTID(mysql.GTIDSet) error {
	return nil
}

func (handler *EventHandler) OnPosSynced(mysql.Position, mysql.GTIDSet, bool) error {
	return nil
}
