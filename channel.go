package mysqlcdc

type Channel struct {
	topic   string
	tables  []string
	handler ChannelHandler
}

type ChannelHandler interface {
	HandleEvent(EventStreamer, *Event)
}

type ChannelHandlerFunc func(EventStreamer, *Event)

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

func (channel *Channel) ChannelHandlerFunc(f func(EventStreamer, *Event)) *Channel {
	return channel.ChannelHandler(ChannelHandlerFunc(f))
}

func (channel *Channel) Match(event *Event) bool {
	for _, table := range channel.tables {
		if table == event.Table.name {
			return true
		}
	}
	return false
}

func (f ChannelHandlerFunc) HandleEvent(sender EventStreamer, event *Event) {
	f(sender, event)
}
