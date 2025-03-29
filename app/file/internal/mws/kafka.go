package mws

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type FileChangeEvent struct {
	EventType string    `json:"event_type"` // "update"
	FileId    int64     `json:"file_id"`
	FolderId  int64     `json:"folder_id"`
	UserId    int32     `json:"user_id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

type KafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaProducer() *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "file-changes",
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
		topic:  "file-changes",
	}
}

func (p *KafkaProducer) SendFileEvent(ctx context.Context, event *FileChangeEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(int(event.UserId))),
		Value: data,
	})
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
