package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"log"
)

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitMQURL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	connection, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %s", err)
	}
	defer connection.Close()

	log.Println("Rabbit MQ connection successful")

	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %s", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatalf("could not subscribe to game logs: %s", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Publishing paused game state")
			if err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			); err != nil {
				log.Printf("could not publish pause state: %v\n", err)
			}
		case "resume":
			fmt.Println("Publishing resume game state")
			if err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			); err != nil {
				log.Printf("could not publish resume game state: %s\n", err)
			}
		case "quit":
			fmt.Println("Stopping...")
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
