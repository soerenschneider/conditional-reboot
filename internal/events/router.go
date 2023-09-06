package events

import (
	"context"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type KafkaEventRouter struct {
	topic     string
	partition int
}

func (r *KafkaEventRouter) New(event event.Event) error {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", r.topic, r.partition)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to dial leader")
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: event.Data()},
	)
	return err
}
