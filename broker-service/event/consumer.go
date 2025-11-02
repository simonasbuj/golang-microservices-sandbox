package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	c := Consumer{
		conn: conn,
	}

	err := c.setup()
	if err != nil {
		return Consumer{}, fmt.Errorf("failed to setup consumer: %w", err)
	}

	return c, nil
}

func (c *Consumer) setup() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	err = declareExchange(channel)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	return nil
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Consumer) Listen(topics []string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return fmt.Errorf("failed to declre random queue: %w", err)
	}

	for _, s := range topics {
		err := ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind a queue: %w", err)
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consume messages: %w", err)
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	log.Printf("waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Printf("error: %s", err)
		}
	default:
		log.Printf("handler for this event doesn't exist: %s", payload)
	}
}

func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service:8072/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request to logger-service: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request to logger-service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("got invalid response, response status: %d", response.StatusCode)
	}

	return nil
}
