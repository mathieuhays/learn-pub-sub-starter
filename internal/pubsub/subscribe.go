package pubsub

import (
	"bytes"
	"encoding/gob"
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

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("error setting prefetch limit: %s", err)
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
			content, err := unmarshaller(msg.Body)
			if err != nil {
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(bytes []byte) (T, error) {
		var content T
		if err := json.Unmarshal(bytes, &content); err != nil {
			return content, err
		}

		return content, nil
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(b []byte) (T, error) {
		var content T
		if err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&content); err != nil {
			return content, err
		}
		return content, nil
	})
}
