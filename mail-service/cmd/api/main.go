package main

import (
	"fmt"
	"log"
	"net/http"
)

type App struct{}

const webPort = "8073"

func main() {
	app := App{}

	log.Println("starting mail service on port", webPort)

	srv := &http.Server{
		Addr: 	fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic("failed to start server: %w", err)
	}
}
