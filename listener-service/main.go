package main

import (
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmqConn, err := connect()
	if err != nil {
		log.Panic(err)
	}
	defer rabbitmqConn.Close()

	log.Println("connected to rabbitmq")

	log.Println("listening for and consuming RabbitMQ messages...")

	consumer, err := event.NewConsumer(rabbitmqConn)
	if err != nil {
		panic(err)
	}

	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Printf("error: %s", err)
	}

}

func connect() (*amqp.Connection, error) {
	connURI := os.Getenv("RABBITMQ_URI")

	var counts int64
	var backOff = 1 * time.Second
	var conn *amqp.Connection

	for {
		c, err := amqp.Dial(connURI)
		if err != nil {
			log.Println("rabbitmq not yet ready, error: %w", err)
			counts++
		} else {
			conn = c
			break
		}

		if counts > 5 {
			log.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Printf("backing off for %s seconds", backOff)
		time.Sleep(backOff)
		continue
	}

	return conn, nil
}