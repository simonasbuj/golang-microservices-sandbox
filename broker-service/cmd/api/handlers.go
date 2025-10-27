package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)


type RequestPayload struct {
	Action 	string 		`json:"action"`
	Auth 	AuthPayload `json:"auth,omitempty"`
	Log 	LogPayload 	`jsong:"log,omitempty"`
}

type AuthPayload struct {
	Email		string	`json:"email"`
	Password	string	`json:"password"`
}

type LogPayload struct {
	Name		string	`json:"name"`
	Data		string	`json:"data"`
}

func (app *App) Broker(w http.ResponseWriter, _ *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "hit the broker, yeah?",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *App) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItem(w, requestPayload.Log)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}

}

func (app *App) authenticate(w http.ResponseWriter, payload AuthPayload) {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")

	// call auth-service
	request, err := http.NewRequest("POST", "http://auth-service:8071/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, errors.New(jsonFromService.Message), http.StatusUnauthorized)
		return
	}

	responsePayload := &jsonResponse{
		Error: false,
		Message: "authenticated successfully",
		Data: jsonFromService.Data,
	}

	app.writeJSON(w, http.StatusAccepted, responsePayload)
}	

func (app *App) logItem(w http.ResponseWriter, log LogPayload) {
	jsonData, _ := json.MarshalIndent(log, "", "\t")

	logServiceUrl := "http://logger-service:8072/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, fmt.Errorf("got invalid response, response status: %d", response.StatusCode))
		return
	}

	res := jsonResponse{
		Error: false,
		Message: "logged",
	}

	app.writeJSON(w, http.StatusAccepted, res)
}