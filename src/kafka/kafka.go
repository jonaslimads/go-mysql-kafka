package kafka

import (
	"bytes"
	"context"
	"fmt"
	mysqlcdc "github.com/jonaslimads/mysql-cdc"
	kafkago "github.com/segmentio/kafka-go"
	"time"
)

type Streamer struct {
	config  *Config
	writers map[string]*kafkago.Writer
}

func NewStreamer(host string, port uint16) Streamer {
	return Streamer{
		config:  &Config{host: host, port: port},
		writers: make(map[string]*kafkago.Writer),
	}
}

// todo add custom data converter
func (streamer Streamer) Stream(topic string, event *mysqlcdc.Event) error {
	buf := new(bytes.Buffer)
	event.BinlogEvent.Dump(buf)

	message := kafkago.Message{
		Topic: topic,
		Key:   []byte("find_a_key"), // todo finish this
		Value: buf.Bytes(),
		Time:  time.Time{},
	}
	return streamer.getWriter(topic).WriteMessages(context.Background(), message)
}

func (streamer Streamer) getWriter(topic string) *kafkago.Writer {
	if writer, ok := streamer.writers[topic]; ok {
		return writer
	}
	return streamer.createWriter(topic)
}

func (streamer Streamer) createWriter(topic string) *kafkago.Writer {
	writer := kafkago.NewWriter(kafkago.WriterConfig{
		Brokers:  []string{streamer.config.getURL()},
		Topic:    topic,
		Balancer: &kafkago.LeastBytes{},
	})
	streamer.writers[topic] = writer
	return writer
}

type Config struct {
	host string
	port uint16
}

func (config Config) getURL() string {
	return fmt.Sprintf("%s:%d", config.host, config.port)
}
