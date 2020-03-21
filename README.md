# go-mysql-kafka

Simple MySQL Change Data Capture that streams changes to Kafka. Built with https://github.com/siddontang/go-mysql.

## Custom handler

You can define handlers to transmit data from any given table to specific Kafka topics:

```go
	eventHandler.HandleChannelFunc("topic", []string{"table1", "table2"}, func(
		streamer kafka.EventStreamer, event *canal.RowsEvent) {
		if err := streamer.Stream("topic", event); err != nil {
			log.Println(err)
		}
	})
```