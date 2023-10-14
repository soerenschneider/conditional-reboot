package checkers

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.uber.org/multierr"
)

const (
	KafkaCheckerName = "kafka"
	defaultGraceTime = 5 * time.Minute
)

type KafkaChecker struct {
	brokers   []string
	topic     string
	partition int
	groupId   string

	mutex       sync.RWMutex
	graceTime   time.Duration
	messageDate time.Time

	acceptedKeys map[string]bool

	reader *kafka.Reader

	certFile string
	keyFile  string
}

type KafkaOpts func(checker *KafkaChecker) error

func NewKafkaChecker(brokers []string, topic string, opts ...KafkaOpts) (*KafkaChecker, error) {
	c := &KafkaChecker{
		brokers:      brokers,
		topic:        topic,
		acceptedKeys: getDefaultAcceptedKeys(),
		graceTime:    defaultGraceTime,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(c); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return c, errs
}

func (c *KafkaChecker) Name() string {
	return KafkaCheckerName
}

func KafkaCheckerFromMap(args map[string]any) (*KafkaChecker, error) {
	// TODO: implement
	panic("not implemented")
}

func getDefaultAcceptedKeys() map[string]bool {
	systemHostname, err := os.Hostname()
	if err != nil {
		log.Warn().Err(err).Msg("could not auto-detect system hostname for kafka's accepted keys")
		return nil
	}

	return map[string]bool{
		systemHostname: true,
	}
}

func (c *KafkaChecker) Start(stop chan bool) error {
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   c.brokers,
		Topic:     c.topic,
		Partition: c.partition,
		MaxBytes:  10e6,
		GroupID:   c.groupId,
	})

	cont := true
	go func() {
		<-stop
		cont = false
		if err := c.reader.Close(); err != nil {
			log.Error().Err(err).Msg("error while closing kafka reader")
		}
	}()

	go func() {
		for cont {
			c.consume()
		}
	}()

	return nil
}

func (c *KafkaChecker) consume() {
	msg, err := c.reader.ReadMessage(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("error while reading message from kafka")
		return
	}

	_, ok := c.acceptedKeys[string(msg.Key)]
	if ok {
		c.mutex.Lock()
		c.messageDate = time.Now()
		c.mutex.Unlock()
	}
}

func (c *KafkaChecker) IsHealthy(ctx context.Context) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// if the timestamp of the message is recent enough we signal we want a reboot
	return time.Since(c.messageDate) <= c.graceTime, nil
}
