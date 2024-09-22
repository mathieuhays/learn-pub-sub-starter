package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

const rabbitMQURL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")

	connection, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %s", err)
	}
	defer connection.Close()
	fmt.Println("Connected to RabbitMQ!")

	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %s", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %s", err)
	}
	gameState := gamelogic.NewGameState(username)

	if err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.SimpleQueueTransient,
		handlerMove(gameState)); err != nil {
		log.Fatalf("could not subscribe to army moves: %s", err)
	}

	if err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState)); err != nil {
		log.Fatalf("could not subscribe to pause: %s", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				move); err != nil {
				log.Printf("error: %s\n", err)
				continue
			}

			fmt.Printf("Moved %v units to %s\n", len(move.Units), move.ToLocation)
		case "spawn":
			if err = gameState.CommandSpawn(words); err != nil {
				fmt.Println(err)
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
