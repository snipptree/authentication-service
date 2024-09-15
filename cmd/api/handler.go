package main

import (
	"authentication/data"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	//log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) SignUp(w http.ResponseWriter, r *http.Request) {
	var user data.User

	// Log the start of the SignUp process
	log.Println("SignUp process started")

	// Step 1: Read JSON input
	log.Println("Reading JSON input")
	err := app.readJSON(w, r, &user)
	if err != nil {
		log.Println("Error reading JSON input:", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println("JSON input successfully read:", user.Email)

	// Step 2: Check if user exists
	log.Println("Checking if user already exists")
	_, err = app.Models.User.GetOne(user.Email)
	if err == mongo.ErrNoDocuments {
		// Step 3: Insert new user
		log.Println("Inserting new user:", user.Email)
		_, err = app.Models.User.Insert(user)
		if err != nil {
			log.Println("Error inserting new user:", err)
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}
		log.Println("New user inserted successfully:", user.Email)
		// Step 4: Log successful sign up
		responsePayload := jsonResponse{
			Error:   false,
			Message: "SignUp Successful",
			Data:    user,
		}
		log.Println("SignUp flow completed successfully for user:", user.Email)
		app.writeJSON(w, http.StatusOK, responsePayload)
	} else {
		log.Println("User already exists:", user.Email)
		responsePayload := jsonResponse{
			Error:   false,
			Message: "Login Successful",
			Data:    user,
		}

		// Log successful login
		log.Println("Logging successful login for user:", user.Email)
		err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
		if err != nil {
			log.Println("Error logging authentication event:", err)
			app.errorJSON(w, err)
			return
		}
		log.Println("Login flow completed successfully for user:", user.Email)
		app.writeJSON(w, http.StatusOK, responsePayload)
		return
	}

	log.Println("Logging sign up event for user:", user.Email)
	err = app.logRequest("authentication", fmt.Sprintf("%s signed up", user.Email))
	if err != nil {
		log.Println("Error logging sign-up event:", err)
		app.errorJSON(w, err)
		return
	}
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceUrl := "http://logger-service/log"
	req, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
