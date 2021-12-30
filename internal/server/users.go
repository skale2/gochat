package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/skale2/gochat/internal/conn"
	"github.com/skale2/gochat/internal/db"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/model"
)

func registerUser(rw http.ResponseWriter, r *http.Request) {
	// Extract new user from request
	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondError(rw, http.StatusBadRequest, fmt.Errorf("invalid request: %v", err))
		return
	}

	// Verify user does not already exist
	_, err = db.GetUser(user.Username)
	if err == nil {
		respondError(rw, http.StatusBadRequest, fmt.Errorf("user %v already exists", user.Username))
		return
	}

	// Create new user in DB
	db.AddUser(&user)

	// Respond with success
	_, err = rw.Write([]byte("successfully created new user"))
	if err != nil {
		log.Error(err)
	} else {
		log.Infof("created new user %v", user.Username)
	}
}

func connectUser(rw http.ResponseWriter, r *http.Request, username model.Username) {
	// Create websocket from http connection and start read/write listeners
	writeChan, err := conn.Create(rw, r, username)
	if err != nil {
		respondError(rw, http.StatusBadRequest, err)
		return
	}

	// Replay unread messages
	messages, err := db.GetUnreadMessages(username)
	if err != nil {
		log.Error(err)
	}

	for _, msg := range messages {
		writeChan <- msg
	}
}
