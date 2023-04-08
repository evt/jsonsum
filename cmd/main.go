package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

const tokenSecret = "secret"

var (
	errEmptyUserNameOrPassword = errors.New("Empty username or password")
	errMethodNotAllowed        = errors.New("Method not allowed")
	errMissingAuthorization    = errors.New("Missing authorization header")
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, errMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var creds Credentials
	err := decoder.Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if creds.Username == "" || creds.Password == "" {
		http.Error(w, errEmptyUserNameOrPassword.Error(), http.StatusBadRequest)
		return
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": creds.Username,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	// signing key.
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", tokenString)
}

func sumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, errMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	// Extract JWT token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, errMissingAuthorization.Error(), http.StatusUnauthorized)
		return
	}
	tokenString := authHeader[len("Bearer "):]

	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the token signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key used to sign the token
		return []byte(tokenSecret), nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Check if the token is valid
	if !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract the subject (username) from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusInternalServerError)
		return
	}
	_, ok = claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid username claim", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var data any
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total := findAndSumNumbers(data)

	sum256 := sha256.Sum256([]byte(total.String()))

	fmt.Fprintf(w, "%x", sum256)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/sum", sumHandler)
	return http.ListenAndServe(":8080", nil)
}
