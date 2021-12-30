package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skale2/gochat/internal/db"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/model"
)

func listThreads(rw http.ResponseWriter, r *http.Request, receipient model.Username) {
	limit, offset, err := db.GetListOpts(r.URL.Query())
	if err != nil {
		respondError(rw, http.StatusBadRequest, err)
	}

	// Retrieve threads from DB
	threads, err := db.ListThreads(model.Username(receipient), limit, offset)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, fmt.Errorf("cannot list threads: %v", err))
		return
	}

	// Marshal threads into list
	threadsBlob, err := json.Marshal(threads)
	if err != nil {
		respondError(rw, http.StatusBadRequest, errors.New("unable to marshal threads into JSON"))
		return
	}

	// Respond with threads
	_, err = rw.Write(threadsBlob)
	if err != nil {
		log.Error(err)
	} else {
		log.Infof("listed threads for user %v, found %v threads", receipient, len(threads))
	}
}

func getThreadMessages(rw http.ResponseWriter, r *http.Request, receipient model.Username) {
	// Retrieve messages from DB
	vars := mux.Vars(r)
	sender := vars["sender"]

	limit, offset, err := db.GetListOpts(r.URL.Query())
	if err != nil {
		respondError(rw, http.StatusBadRequest, err)
	}

	msgs, err := db.ListThreadMessages(receipient, model.Username(sender), limit, offset)
	if err != nil {
		respondError(rw, http.StatusInternalServerError, fmt.Errorf("cannot list thread messages: %v", err))
		return
	}

	// Marshal messages into list
	msgsBlob, err := json.Marshal(msgs)
	if err != nil {
		respondError(rw, http.StatusBadRequest, errors.New("unable to marshal thread messages into JSON"))
		return
	}

	// Respond with messages
	_, err = rw.Write(msgsBlob)
	if err != nil {
		log.Error(err)
	} else {
		log.Infof("listed thread messages sent by %v to %v, found %v messages", sender, receipient, len(msgs))
	}
}

func readThread(rw http.ResponseWriter, r *http.Request, receipient model.Username) {
	vars := mux.Vars(r)
	sender := vars["sender"]

	// Mark thread as read in DB
	err := db.ReadThread(receipient, model.Username(sender))
	if err != nil {
		respondError(rw, http.StatusInternalServerError, fmt.Errorf("cannot mark thread as read: %v", err))
		return
	}

	// Respond with success
	_, err = rw.Write([]byte("successfully marked thread as read"))
	if err != nil {
		log.Error(err)
	} else {
		log.Infof("marked thread sent by %v to %v as read", sender, receipient)
	}
}
