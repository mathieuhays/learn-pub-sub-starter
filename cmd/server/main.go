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
		log.Fatal(err)
	}
	defer connection.Close()

	log.Println("Rabbit MQ connection successful")

	mqChan, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	if err = pubsub.PublishJSON(
		mqChan,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	); err != nil {
		log.Fatal(err)
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		if words[0] == "pause" {
			fmt.Println("sending pause message...")
			if err = pubsub.PublishJSON(
				mqChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			); err != nil {
				log.Fatal(err)
			}
			continue
		}

		if words[0] == "resume" {
			fmt.Println("sending resume message...")
			if err = pubsub.PublishJSON(
				mqChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			); err != nil {
				log.Fatal(err)
			}
			continue
		}

		if words[0] == "quit" {
			fmt.Println("Stopping...")
			break
		}

		fmt.Println("I don't understand. Either use 'pause', 'resume' or 'quit'")
	}

	//// Wait for interrupt signal
	//signalChan := make(chan os.Signal, 1)
	//signal.Notify(signalChan, os.Interrupt)
	//<-signalChan

	log.Println("Shutting Down")
}
