package main

import (
	"log"
	"os"
)

func main() {
	eventStreamer := NewEventStreamer(NewMySqlConfigFromEnv())

	eventStreamer.HandleStreamFunc("employees", "employees", func(event *Event) {
		if event.action == UpdateAction {
			log.Println("got employee update binlogEvent", event.table, event.action)
			event.binlogEvent.Dump(os.Stdout)
		}
	})

	eventStreamer.HandleStreamFunc("employees", "titles", func(event *Event) {
		if event.action == UpdateAction {
			log.Println("got title update binlogEvent", event.table, event.action)
			event.binlogEvent.Dump(os.Stdout)
		}
	})

	eventStreamer.run()
}
