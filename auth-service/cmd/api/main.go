package main

import (
	"auth-service/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "8071"

var connectionCounts int

type App struct {
	DB 		*sql.DB
	Models	data.Models
}

func main() {
	log.Printf("starting auth-service at port: %s", webPort)

	// TODO connect to DB
	conn := connectToDb()
	if conn == nil {
		log.Panic("can't connect to postgres")
	}

	// start app
	app := App{
		DB:		conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr: 		fmt.Sprintf(":%s", webPort),
		Handler:	app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDb() *sql.DB {
	dsn := os.Getenv("DSM")

	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet eady...")
			connectionCounts++
		} else {
			log.Println("connected to postgres")
			return conn
		}

		if connectionCounts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("backing off for 2 seconds...")
		time.Sleep(time.Second * 2)
		continue
	}
}

