package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type App struct{
	Mailer	Mail
}

const webPort = "8073"

func main() {
	app := App{
		Mailer: createMail(),
	}

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

func createMail() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := Mail{
		Domain: 	os.Getenv("MAIL_DOMAIN"),
		Host: 		os.Getenv("MAIL_HOST"),
		Port: 		port,
		Username: 	os.Getenv("MAIL_USERNAME"),
		Password: 	os.Getenv("MAIL_PASSWORD"),
		Encryption: os.Getenv("MAIL_ENCRYPTION"),
		FromName: 	os.Getenv("MAIL_FROM_NAME"),
		FromAddres: os.Getenv("MAIL_FROM_ADDRESS"),
	}
	
	return m
}
