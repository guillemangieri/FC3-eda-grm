package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	producer *ckafka.Producer
	timeout  time.Duration
}

func NewKafkaProducer(configMap *ckafka.ConfigMap) (*Producer, error) {
	producer, err := ckafka.NewProducer(configMap)
	if err != nil {
		return nil, err
	}
	return &Producer{producer: producer, timeout: 10 * time.Second}, nil
}

func (p *Producer) Publish(msg interface{}, key []byte, topic string) error {
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	message := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &topic, Partition: ckafka.PartitionAny},
		Value:          msgJson,
		Key:            key,
	}
	delivery := make(chan ckafka.Event, 1)
	if err := p.producer.Produce(message, delivery); err != nil {
		return err
	}

	select {
	case event := <-delivery:
		delivered, ok := event.(*ckafka.Message)
		if !ok {
			return fmt.Errorf("unexpected Kafka delivery event %T", event)
		}
		if delivered.TopicPartition.Error != nil {
			return delivered.TopicPartition.Error
		}
		return nil
	case <-time.After(p.timeout):
		return fmt.Errorf("Kafka delivery confirmation timed out after %s", p.timeout)
	}
}

func (p *Producer) Close() {
	if p != nil && p.producer != nil {
		p.producer.Flush(int(p.timeout.Milliseconds()))
		p.producer.Close()
	}
}
