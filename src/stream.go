package main

type Stream struct {
	topic         string
	path          string
	streamHandler StreamHandler
}

type StreamHandler interface {
	Stream(*Event)
}

type StreamHandlerFunc func(*Event)

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

func (stream *Stream) Match(event *Event) bool {
	return stream.path == event.table.name
}

func (streamHandlerFunc StreamHandlerFunc) Stream(event *Event) {
	streamHandlerFunc(event)
}
