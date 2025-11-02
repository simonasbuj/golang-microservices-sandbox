// Package main
package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "8070"

type App struct {
	rabbitmq *amqp.Connection
}

func main() {
	rabbitmqConn, err := connect()
	if err != nil {
		log.Panic(err)
	}
	defer rabbitmqConn.Close()

	log.Println("connected to rabbitmq")

	app := App{
		rabbitmq: rabbitmqConn,
	}

	log.Printf("starting broker service on port %s", webPort)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", webPort),
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connect() (*amqp.Connection, error) {
	connURI := os.Getenv("RABBITMQ_URI")

	var counts int64
	var conn *amqp.Connection

	for {
		c, err := amqp.Dial(connURI)
		if err != nil {
			log.Printf("rabbitmq not yet ready, error: %s", err)
			counts++
		} else {
			conn = c
			break
		}

		if counts > 5 {
			log.Println(err)
			return nil, err
		}

		backOff := time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Printf("backing off for %s seconds", backOff)
		time.Sleep(backOff)
		continue
	}

	return conn, nil
}
