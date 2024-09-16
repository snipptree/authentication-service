package main

import (
	"authentication/data"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// Secret key for signing JWT tokens
var jwtSecret = []byte("REDACTED_JWT_SECRET")

func (app *Config) GenerateJWT(user data.User) (string, error) {
	claims := jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(180 * time.Hour * 24).Unix(), // Token expiration time
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token's signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrAbortHandler
		}
		return jwtSecret, nil
	})
	return token, err
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
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
	_, err = app.Repo.GetOne(user.Email)
	if err == mongo.ErrNoDocuments {
		// Step 3: Insert new user
		log.Println("Inserting new user:", user.Email)
		_, err = app.Repo.Insert(user)
		if err != nil {
			log.Println("Error inserting new user:", err)
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}
		log.Println("New user inserted successfully:", user.Email)

		// Step 4: Generate JWT token
		token, err := app.GenerateJWT(user)
		if err != nil {
			log.Println("Error generating JWT token:", err)
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}

		// Step 5: Log successful sign up and send response with JWT
		responsePayload := jsonResponse{
			Error:   false,
			Message: "SignUp Successful",
			Data: map[string]interface{}{
				"user":  user,
				"token": token,
			},
		}
		log.Println("SignUp flow completed successfully for user:", user.Email)
		app.writeJSON(w, http.StatusOK, responsePayload)
		return
	} else {
		log.Println("User already exists:", user.Email)

		// Step 4: Generate JWT token
		token, err := app.GenerateJWT(user)
		if err != nil {
			log.Println("Error generating JWT token:", err)
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}

		// Step 5: Log successful login and send response with JWT
		responsePayload := jsonResponse{
			Error:   false,
			Message: "Login Successful",
			Data: map[string]interface{}{
				"user":  user,
				"token": token,
			},
		}
		log.Println("Login flow completed successfully for user:", user.Email)
		app.writeJSON(w, http.StatusOK, responsePayload)
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
