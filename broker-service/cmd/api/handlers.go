package main

import (
	"net/http"
)

func (app *App) Broker(w http.ResponseWriter, _ *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "hit the broker, yeah?",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}
