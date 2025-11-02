package main

import (
	"broker-service/event"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
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
		app.logEventViaRPC(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *App) authenticate(w http.ResponseWriter, payload AuthPayload) {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")

	// call auth-service
	request, err := http.NewRequest(
		"POST",
		"http://auth-service:8071/authenticate",
		bytes.NewBuffer(jsonData),
	)
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
		Error:   false,
		Message: "authenticated successfully",
		Data:    jsonFromService.Data,
	}

	app.writeJSON(w, http.StatusAccepted, responsePayload)
}

//nolint:unused
func (app *App) logItem(w http.ResponseWriter, log LogPayload) {
	jsonData, _ := json.MarshalIndent(log, "", "\t")

	logServiceURL := "http://logger-service:8072/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
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
		app.errorJSON(
			w,
			fmt.Errorf("got invalid response, response status: %d", response.StatusCode),
		)
		return
	}

	res := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJSON(w, http.StatusAccepted, res)
}

func (app *App) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	mailServiceURL := "http://mail-service:8073/send"

	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
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

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("error reading response body: %v", err))
		return
	}

	bodyStr := string(bodyBytes)

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(
			w,
			fmt.Errorf(
				"error from mail-service, got status code: %d, response: %s",
				response.StatusCode,
				bodyStr,
			),
		)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Message sent to " + msg.To,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

//nolint:unused
func (app *App) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		log.Println(err)
	}

	payload := jsonResponse{
		Error:   false,
		Message: "logged via RabbitMQ",
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

//nolint:unused
func (app *App) pushToQueue(name, message string) error {
	emitter, err := event.NewEventEmmiter(app.rabbitmq)
	if err != nil {
		return fmt.Errorf("failed to create new Emitter: %w", err)
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.Marshal(&payload)
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return fmt.Errorf("failed to push event to rabbitmq: %w", err)
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

//nolint:unused
func (app *App) logEventViaRPC(w http.ResponseWriter, l LogPayload) {
	log.Println("logging event via RCP")

	client, err := rpc.Dial("tcp", "logger-service:7072")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload(l)

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: result,
	})
}
