package produser

import (
	"time"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-faster/errors"
	"github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(kafkaBrokers string) (*KafkaProducer, error) {

	//brokers := strings.Join(cfg.KafkaBrokers, ",")

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
	})
	if err != nil {
		logrus.Errorf("connecting to kafka: %#v", err)
		return &KafkaProducer{}, errors.Wrap(err, "failed to connect to kafka")
	}

	logrus.Infof("connected to kafka brokers: %s", kafkaBrokers)

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logrus.Errorf("delivery failed: %v", ev.TopicPartition)
				} else {
					logrus.Infof("message delivered to %v", ev.TopicPartition)
				}
			}
		}
	}()
	return &KafkaProducer{
		producer: producer,
	}, nil
}

func (kp *KafkaProducer) SendMessage(topic string, key, value []byte, createdAt time.Time) error {

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
		Timestamp:      createdAt,
	}

	err := kp.producer.Produce(msg, nil)
	if err != nil {
		return errors.Wrap(err, "failed to produce message")
	}

	return nil
}

func (kp *KafkaProducer) Close() {
	kp.producer.Flush(5000)
	kp.producer.Close()
	logrus.Infof("kafka producer closed")
}
