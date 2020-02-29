package main

type Stream struct {
	topic         string
	tables        []string
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

func (stream *Stream) Tables(tables []string) *Stream {
	stream.tables = tables
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
	for _, table := range stream.tables {
		if table == event.table.name {
			return true
		}
	}
	return false
}

func (streamHandlerFunc StreamHandlerFunc) Stream(event *Event) {
	streamHandlerFunc(event)
}
