package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/skale2/gochat/internal/conn"
	"github.com/skale2/gochat/internal/db"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/model"
)

const HttpPort = 8080

func SendErrorResponse(rw http.ResponseWriter, status int, err error) {
	log.Error(err)
	rw.WriteHeader(status)
	rw.Write([]byte(err.Error()))
}

func initialize(ctx context.Context) {
	conn.Initialize()
	log.Initialize(log.FileWriter())
	db.Initialize()
}

func finalize(sigintChan chan<- os.Signal, cancel context.CancelFunc) {
	signal.Stop(sigintChan)
	cancel()
	db.Finalize()
}

func authenticate(rw http.ResponseWriter, r *http.Request) (model.Username, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return "", errors.New("unable to extract authentication credentials from header")
	}

	user, err := db.GetUser(model.Username(username))
	if err != nil {
		return "", fmt.Errorf("unable to find user %v", username)
	}

	if user.Password != password {
		return "", fmt.Errorf("invalid username/password combination")
	}

	log.Infof("successfully authenticated user %v", username)
	return model.Username(username), nil
}

func replayUnread(username model.Username, writeInp chan<- *model.Message) {
	messages, err := db.GetUnreadMessages(username)
	if err != nil {
		log.Error(err)
	}

	for _, msg := range messages {
		writeInp <- msg
	}
}

func registerHandler(rw http.ResponseWriter, r *http.Request) {
	// Extract new user from request
	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		SendErrorResponse(rw, 400, fmt.Errorf("invalid request: %v", err))
		return
	}

	// Verify user does not already exist
	_, err = db.GetUser(user.Username)
	if err == nil {
		SendErrorResponse(rw, 400, fmt.Errorf("user %v already exists", user.Username))
		return
	}

	// Create new user in DB
	db.AddUser(&user)

	rw.Write([]byte("successfully created new user"))

	log.Infof("created new user %v", user.Username)
}

func connectHandler(rw http.ResponseWriter, r *http.Request) {
	// Authenticate
	username, err := authenticate(rw, r)
	if err != nil {
		SendErrorResponse(rw, 401, err)
		return
	}

	// Create websocket from http connection and start read/write listeners
	writeChan, err := conn.Create(rw, r, username)
	if err != nil {
		SendErrorResponse(rw, 400, err)
		return
	}

	// Replay unread messages
	replayUnread(username, writeChan)
}

func main() {
	// Create global application context
	ctx, cancel := context.WithCancel(context.Background())

	// Create handler to cancel context on interrupt signal
	sigintChan := make(chan os.Signal, 1)
	signal.Notify(sigintChan, os.Interrupt)

	defer func() {
		defer finalize(sigintChan, cancel)
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	// Initialize application
	initialize(ctx)

	// Setup HTTP server handlers
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/connect", connectHandler)

	log.Infof("Starting server on port %v", HttpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%v", HttpPort), nil)
	if err != nil {
		log.Error(err)
	}
}
