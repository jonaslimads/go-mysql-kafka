package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/siddontang/go-mysql/canal"
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

// todo lower case keys from event
func (streamer Streamer) Stream(topic string, event *canal.RowsEvent) error {
	key := fmt.Sprintf("%s-%s-%d", event.Table.String(), event.Action, event.Header.Timestamp)
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
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
