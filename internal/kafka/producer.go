package kafka

import (
	"banner-rotation/internal/pkg/events"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// WriterInterface определяет контракт для работы с Kafka
type WriterInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type ProducerInterface interface {
	SendEvent(ctx context.Context, eventType events.EventType, slotID, bannerID, groupID int) error
	Close() error
}

type Producer struct {
	writer WriterInterface
}

func NewProducer(writer WriterInterface) *Producer {
	return &Producer{writer: writer}
}

func (p *Producer) SendEvent(ctx context.Context, eventType events.EventType, slotID, bannerID, groupID int) error {
	event := events.BannerEvent{
		Type:      eventType,
		SlotID:    slotID,
		BannerID:  bannerID,
		GroupID:   groupID,
		Timestamp: time.Now().UTC(),
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Value: jsonData,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
