package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type RequestError struct {
	Field string
	Tag   string
	Value string
}

func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	requestErrors := validateRequestPayload(data)
	if len(requestErrors) != 0 {
		return getFirstError(data, requestErrors)
	}
	return nil
}

func getFirstError(data any, requestErrors []*RequestError) error {
	firstError := *requestErrors[0]
	//jsonFieldName, _ := reflect.TypeOf(data).FieldByName(firstError.Field)
	errorMessage := fmt.Sprintf("%v %v %v", firstError.Field, firstError.Tag, firstError.Value)

	return errors.New(errorMessage)
}

func validateRequestPayload(data any) []*RequestError {
	var validate *validator.Validate
	validate = validator.New()

	err := validate.Struct(data)

	var requestErrors []*RequestError

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var el RequestError
			el.Field = err.Field()
			el.Tag = err.Tag()
			el.Value = err.Param()
			requestErrors = append(requestErrors, &el)
		}
	}

	return requestErrors
}

func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}

func (app *Config) addCookies(w http.ResponseWriter, cookies ...*http.Cookie) {
	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}
}

func (app *Config) validateToken(email string, token string) (any, error) {
	requestPaylod := AuthRequest{
		Email: email,
		Token: token,
	}
	jsonData, _ := json.MarshalIndent(requestPaylod, "", "\t")

	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error while creating auth request %s", err)
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Got error from auth service %s", err)
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		log.Println("Got unauthorized error from auth service")
		return nil, errors.New("invalid session")
	}

	var session JsonResponse

	return session, nil
}

func (app *Config) GenerateToken(email string) (string, error) {

	var requestPayload = AuthRequest{
		Email: email,
	}

	jsonData, _ := json.MarshalIndent(requestPayload, "", "\t")

	request, err := http.NewRequest("GET", "http://authentication-service/token", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error while creating auth request %s", err)
		return "payload", err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Got error from auth service %s", err)
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("Got error from auth service")
		return "", errors.New("invalid session")
	}

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return "", errors.New("error reading response body")
	}

	var responsePayload JsonResponse
	err = json.Unmarshal(body, &responsePayload)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return "", err
	}

	authResponse, ok := responsePayload.Data.(map[string]any)
	if !ok {
		return "", fmt.Errorf("data is not a map[string]interface{}")
	}

	result := make(map[string]string)
	for key, value := range authResponse {
		// Perform a type assertion to convert value to string
		strValue, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("value for key '%s' is not a string", key)
		}
		result[key] = strValue
	}

	return result["token"], nil
}

func (app *Config) revokeSession(email string) error {
	var requestPayload = AuthRequest{
		Email: email,
	}

	jsonRequestData, _ := json.MarshalIndent(requestPayload, "", "\t")
	request, err := http.NewRequest("DELETE", "http://authentication-service/revoke", bytes.NewBuffer(jsonRequestData))
	if err != nil {
		log.Printf("error in making request, %s", err)
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("error while sending request, %s", err)
		return err
	}

	if response.StatusCode != http.StatusAccepted {
		log.Printf("Got error from auth service, %s", err)
		return err
	}

	log.Printf("[User=%s] Session revoked. Bye Bye !!", email)
	return nil
}
