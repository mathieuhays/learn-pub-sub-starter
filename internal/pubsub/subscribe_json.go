package pubsub

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			var content T
			if err = json.Unmarshal(msg.Body, &content); err != nil {
				fmt.Printf("could not unmarshall message: %v\n", err)
				continue
			}

			switch handler(content) {
			case Ack:
				fmt.Println("Ack")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("NackRequeue")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("NackDiscard")
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}
