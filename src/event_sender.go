package main

type EventSender interface {
	Send(topic string, event *Event)
	//SendCustom(topic string, event *Event)
}
