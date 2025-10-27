package main

import (
	"logger-service/data"
	"net/http"
)

type jsonPayload struct {
	Name	string	`json:"name"`
	Data	string	`json:"data"`
}

func (app *App) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload jsonPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	logEntry := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err = app.Models.LogEntry.Insert(logEntry)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	resp := jsonResponse{
		Error: false,
		Message: "log entry saved",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}
