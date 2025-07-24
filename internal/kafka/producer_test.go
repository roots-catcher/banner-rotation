package kafka

import (
	"banner-rotation/internal/pkg/events"
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type MockWriter struct {
	messages []kafka.Message
}

func (m *MockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *MockWriter) Close() error {
	return nil
}

func TestProducer(t *testing.T) {
	mockWriter := &MockWriter{}
	producer := &Producer{writer: mockWriter}

	t.Run("SendEvent success", func(t *testing.T) {
		err := producer.SendEvent(context.Background(), events.EventShow, 1, 2, 3)
		assert.NoError(t, err)
		assert.Len(t, mockWriter.messages, 1)
	})

	t.Run("Close success", func(t *testing.T) {
		err := producer.Close()
		assert.NoError(t, err)
	})
}
