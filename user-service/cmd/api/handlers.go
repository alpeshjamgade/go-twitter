package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"user-service/data"
)

func (app *Config) Signup(w http.ResponseWriter, r *http.Request) {
	var requestPayload data.User
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Print(err)
		app.errorJSON(w, errors.New(fmt.Sprintf("Error while reading request. Error : %s", err)), http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.Insert(requestPayload)
	if err != nil {
		log.Print(err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("User created"),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Login(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Print(err)
		app.errorJSON(w, errors.New(fmt.Sprintf("Error while reading request. Error : %s", err)), http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		log.Print(err)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	match, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !match {
		if err != nil {
			log.Printf("Error while matching password. %v", err)
		}

		app.errorJSON(w, errors.New("email or Password doesnt match"), http.StatusBadRequest)
		return
	}

	tokenResponse, err := app.GenerateToken(requestPayload.Email)
	if err != nil {
		log.Printf("error from authentication service, %s", err)
		app.errorJSON(w, err, http.StatusForbidden)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    tokenResponse,
	}

	app.addCookies(
		w,
		&http.Cookie{Name: "email", Value: requestPayload.Email, Expires: time.Now().Add(24 * time.Hour), HttpOnly: true},
		&http.Cookie{Name: "Authorization", Value: tokenResponse, Expires: time.Now().Add(24 * time.Hour), HttpOnly: true},
	)
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) UserProfile(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, errors.New(fmt.Sprintf("Error while reading request. Error : %s", err)), http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		payload := JsonResponse{
			Error:   true,
			Message: "user not found",
			Data:    map[string]string{},
		}
		app.writeJSON(w, http.StatusNotFound, payload)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: "user profile fetched successfully",
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Print(err)
		app.errorJSON(w, errors.New(fmt.Sprintf("Error while reading request. Error : %s", err)), http.StatusBadRequest)
		return
	}

	log.Printf("Logging out user %s", requestPayload.Email)

	if err = app.revokeSession(requestPayload.Email); err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: "user logged out successfully",
		Data:    map[string]string{},
	}

	app.addCookies(
		w,
		&http.Cookie{Name: "email", Value: "", Expires: time.Now().Add(-1), HttpOnly: true},
		&http.Cookie{Name: "Authorization", Value: "", Expires: time.Now().Add(-1), HttpOnly: true},
	)
	app.writeJSON(w, http.StatusAccepted, payload)

}
