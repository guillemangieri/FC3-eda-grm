package handler

import (
	"fmt"

	"github.com.br/devfullcycle/fc-ms-wallet/pkg/events"
	"github.com.br/devfullcycle/fc-ms-wallet/pkg/kafka"
)

type TransactionCreatedKafkaHandler struct {
	Kafka *kafka.Producer
}

func NewTransactionCreatedKafkaHandler(kafka *kafka.Producer) *TransactionCreatedKafkaHandler {
	return &TransactionCreatedKafkaHandler{
		Kafka: kafka,
	}
}

func (h *TransactionCreatedKafkaHandler) Handle(message events.EventInterface) error {
	if err := h.Kafka.Publish(message, nil, "transactions"); err != nil {
		return err
	}
	fmt.Println("TransactionCreatedKafkaHandler: ", message.GetPayload())
	return nil
}
