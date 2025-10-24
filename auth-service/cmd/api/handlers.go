package main

import (
	"errors"
	"fmt"
	"net/http"
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

	payload := jsonResponse{
		Error: false,
		Message: fmt.Sprintf("logged in user %s", user.Email),
		Data: user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
