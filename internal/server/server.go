package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skale2/gochat/internal/db"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/model"
)

func Create(ctx context.Context, cancel context.CancelFunc, port int) *http.Server {
	// Create http handler with routes
	r := mux.NewRouter()

	r.HandleFunc("/register", registerUser).Methods(http.MethodPost)
	r.HandleFunc("/connect", authThenServe(connectUser))
	r.HandleFunc("/threads", authThenServe(listThreads)).Methods(http.MethodGet)
	r.HandleFunc("/threads/{sender}", authThenServe(getThreadMessages)).Methods(http.MethodGet)
	r.HandleFunc("/threads/{sender}/read", authThenServe(readThread)).Methods(http.MethodPost)

	// Create http server
	server := &http.Server{
		Addr:        fmt.Sprintf(":%v", port),
		Handler:     r,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	// server.RegisterOnShutdown(cancel)
	return server
}

func Start(server *http.Server) {
	// Start http server
	msg := fmt.Sprintf("Starting server at %v...\n", server.Addr)
	fmt.Print(msg)
	log.Info(msg)

	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		msg = "Shutting down server..."
		fmt.Print(msg)
		log.Info(msg)
	} else {
		log.Error(err)
	}
}

func Shutdown(ctx context.Context, server *http.Server) {
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}

func respondError(rw http.ResponseWriter, status int, err error) {
	log.Error(err)
	rw.WriteHeader(status)
	rw.Write([]byte(err.Error()))
}

type authHandler func(http.ResponseWriter, *http.Request, model.Username)

func authThenServe(handler authHandler) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		username, err := auth(r)
		if err != nil {
			respondError(rw, http.StatusUnauthorized, err)
			return
		}

		handler(rw, r, username)
	}
}

func auth(r *http.Request) (model.Username, error) {
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
