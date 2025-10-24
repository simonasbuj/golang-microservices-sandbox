package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)


type RequestPayload struct {
	Action 	string 		`json:"action"`
	Auth 	AuthPayload `json:"auth,omitempty`
}

type AuthPayload struct {
	Email		string	`json:"email"`
	Password	string	`json:"password"`
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
