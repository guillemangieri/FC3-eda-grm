package balances

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

const Topic = "balances"

func Consume(ctx context.Context, db *sql.DB, brokers, group string) error {
	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"group.id":                 group,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
	})
	if err != nil {
		return err
	}
	defer consumer.Close()
	// Wallet publishes without a key/version. One partition preserves broker order.
	metadata, err := consumer.GetMetadata(ptr(Topic), false, 10000)
	if err != nil {
		return err
	}
	info, ok := metadata.Topics[Topic]
	if !ok || info.Error.Code() != ckafka.ErrNoError || len(info.Partitions) != 1 {
		return fmt.Errorf("topic %s must have exactly one partition", Topic)
	}
	if err := consumer.SubscribeTopics([]string{Topic}, nil); err != nil {
		return err
	}
	log.Printf("Balances consuming topic=%s group=%s", Topic, group)
	for ctx.Err() == nil {
		message, err := consumer.ReadMessage(time.Second)
		if err != nil {
			if kafkaErr, ok := err.(ckafka.Error); ok && kafkaErr.Code() == ckafka.ErrTimedOut {
				continue
			}
			log.Printf("Kafka read failed: %v", err)
			if !pause(ctx) {
				break
			}
			continue
		}
		// Do not advance past a failed record. Invalid events also block visibly.
		processingStarted := time.Now()
		for ctx.Err() == nil {
			event, err := Decode(message.Value)
			if err == nil {
				var applied bool
				applied, err = Apply(ctx, db, Topic, message.TopicPartition.Partition, int64(message.TopicPartition.Offset), event)
				if err == nil {
					log.Printf("Persisted partition=%d offset=%v applied=%t", message.TopicPartition.Partition, message.TopicPartition.Offset, applied)
					break
				}
			}
			log.Printf("Processing failed partition=%d offset=%v; retrying without commit: %v", message.TopicPartition.Partition, message.TopicPartition.Offset, err)
			// Rejoin before max.poll.interval.ms expires during a prolonged outage.
			if time.Since(processingStarted) > 2*time.Minute {
				return fmt.Errorf("processing retry window exhausted: %w", err)
			}
			if !pause(ctx) {
				break
			}
		}
		if ctx.Err() == nil {
			if _, err := consumer.CommitMessage(message); err != nil {
				// Restart/rejoin rather than looping on a revoked group assignment.
				return fmt.Errorf("Kafka commit failed; restart will replay safely: %w", err)
			}
		}
	}
	return ctx.Err()
}

func ptr(value string) *string { return &value }

func pause(ctx context.Context) bool {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
