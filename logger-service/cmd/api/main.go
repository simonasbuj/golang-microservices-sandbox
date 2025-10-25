package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "8072"
	rpcPort  = "7072"
	mongoURL = "mongodb://mongo:27017"
	grpcPort = "9072"
)

var client *mongo.Client

type App struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := App{
		Models: data.New(client),
	}

	app.serve()
}

func (app *App) serve() {
	log.Printf("starting logger service on port %s", webPort)

	srv := &http.Server{
		Addr: 				fmt.Sprintf(":%s", webPort),
		Handler:           	app.Routes(),
		ReadHeaderTimeout: 	5 * time.Second,
	}
	
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admig",
		Password: "password",
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("error connecting: ", err)
	}

	return c, nil
}