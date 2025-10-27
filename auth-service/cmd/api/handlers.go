package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func (app *App) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email 	string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// validate user against db
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	ok, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !ok {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	// log auth request
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		// print fail, but continue, cause it's not a deal breaker
		log.Println("error: failed to log authentication")
	}

	payload := jsonResponse{
		Error: false,
		Message: fmt.Sprintf("logged in user %s", user.Email),
		Data: user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}


func (app *App) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceUrl := "http://logger-service:8072/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request to logger-service, %w", err)
	}
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request to logger-service, %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logger service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}