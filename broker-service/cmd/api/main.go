// Package main
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const webPort = "8070"

type App struct{}

func main() {
	app := App{}

	log.Printf("starting broker service on port %s", webPort)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", webPort),
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
