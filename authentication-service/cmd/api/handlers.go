package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email string `json:"user"`
		Token string `json:"token"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Printf("error while reading response %s", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	var cachedToken string
	jsonToken := app.Cache.HGet("userTokens", requestPayload.Email)

	err = json.Unmarshal([]byte(jsonToken), &cachedToken)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	if cachedToken != requestPayload.Token {
		log.Print("invalid token or token doesnt match")
		app.errorJSON(w, errors.New("Request unauthorized"), http.StatusUnauthorized)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Valid token for user %s", requestPayload.Email),
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Printf("error while reading response %s", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	token := "authToken"
	_, err = app.Cache.HSetNX("userTokens", requestPayload.Email, token, 24*time.Hour)
	if err != nil {
		log.Printf("error while setting redis key, %s", err)
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Token generated for %s", requestPayload.Email),
		Data: map[string]string{
			"email": requestPayload.Email,
			"token": token,
		},
	}

	app.writeJSON(w, http.StatusAccepted, &payload)

}

func (app *Config) RevokeSession(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Printf("error while reading response %s", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	app.Cache.HDel("userTokens", requestPayload.Email)

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("session revoked for %s", requestPayload.Email),
		Data:    map[string]string{},
	}

	app.writeJSON(w, http.StatusAccepted, &payload)

}
