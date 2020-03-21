package kafka

import "github.com/siddontang/go-mysql/canal"

type Channel struct {
	topic   string
	tables  []string
	handler ChannelHandler
}

type ChannelHandler interface {
	HandleEvent(EventStreamer, *canal.RowsEvent)
}

type ChannelHandlerFunc func(EventStreamer, *canal.RowsEvent)

func (channel *Channel) Topic(topic string) *Channel {
	channel.topic = topic
	return channel
}

func (channel *Channel) Tables(tables []string) *Channel {
	channel.tables = tables
	return channel
}

func (channel *Channel) ChannelHandler(handler ChannelHandler) *Channel {
	channel.handler = handler
	return channel
}

func (channel *Channel) ChannelHandlerFunc(f func(EventStreamer, *canal.RowsEvent)) *Channel {
	return channel.ChannelHandler(ChannelHandlerFunc(f))
}

func (channel *Channel) Match(event *canal.RowsEvent) bool {
	for _, table := range channel.tables {
		if table == event.Table.Name {
			return true
		}
	}
	return false
}

func (f ChannelHandlerFunc) HandleEvent(sender EventStreamer, event *canal.RowsEvent) {
	f(sender, event)
}
