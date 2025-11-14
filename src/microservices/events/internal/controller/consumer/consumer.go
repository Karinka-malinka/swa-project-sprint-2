package consumer

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-faster/errors"
	"github.com/sirupsen/logrus"
)

type KafkaConsumer struct {
	Consumer *kafka.Consumer
}

func NewKafkaConsumer(kafkaBrokers, topic string) (*KafkaConsumer, error) {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaBrokers,
		"group.id":           "events",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error in connecting kafka consumer")
	}

	err = consumer.Subscribe(topic, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to subscribe to topic %s", topic)
	}

	return &KafkaConsumer{
		Consumer: consumer,
	}, nil
}

func (kc *KafkaConsumer) ListenForMessages(ctx context.Context, timeout time.Duration) error {

	defer kc.Consumer.Close()
	logrus.Infof("Kafka consumer started")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ev := kc.Consumer.Poll(int(10 * timeout.Seconds()))
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				logrus.Infof("message received: topic=%s partition=%d offset=%d key=%s",
					*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset, string(e.Key))
				logrus.Infof("message value: %s", e.Value)
			case kafka.Error:
				logrus.Errorf("kafka error: %v", e)
				return e
			}
		}
	}
}
