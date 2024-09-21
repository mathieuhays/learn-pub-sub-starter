package pubsub

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Durable = iota
	Transient
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        b,
	})
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	mqChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := true
	autoDelete := false
	exclusive := false

	if simpleQueueType == Transient {
		durable = false
		autoDelete = true
		exclusive = true
	}

	queue, err := mqChan.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		mqChan.Close()
		return nil, amqp.Queue{}, err
	}

	err = mqChan.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		mqChan.Close()
		return nil, amqp.Queue{}, err
	}

	return mqChan, queue, nil
}
