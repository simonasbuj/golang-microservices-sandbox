package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	webPort  = "8072"
	rpcPort  = "7072"
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

	err = rpc.Register(new(RPCServer))
	if err != nil {
		log.Printf("failed to register rcp server: %s", err)
	}
	go app.rpcListen()
	
	go app.gRPCListen()

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
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL"))
	clientOptions.SetAuth(options.Credential{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("error connecting: ", err)
	}

	err = c.Ping(context.Background(), &readpref.ReadPref{})
	if err != nil {
		log.Panic("error pinging mongodb: %w", err)
	}

	log.Println("connected to mongodb")

	return c, nil
}

func (app *App) rpcListen() error {
	log.Println("starting rpc server on port: ", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		log.Printf("failed to start rcp server: %s", err)
		return fmt.Errorf("failed to start rcp server: %w", err)
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConn)
	}
}
