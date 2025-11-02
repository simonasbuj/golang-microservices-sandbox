package event

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emmitter struct {
	rabbitmq *amqp.Connection
}

func (e *Emmitter) setup() error {
	channel, err := e.rabbitmq.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer channel.Close()

	return declareExchange(channel)
}

func (e *Emmitter) Push(event, severity string) error {
	channel, err := e.rabbitmq.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer channel.Close()

	log.Println("pushing to channel")

	err = channel.Publish(
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte(event),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func NewEventEmmiter(conn *amqp.Connection) (*Emmitter, error) {
	emmitter := Emmitter{
		rabbitmq: conn,
	}

	err := emmitter.setup()
	if err != nil {
		return nil, fmt.Errorf("failed to setup emitter")
	}

	return &emmitter, nil
}
