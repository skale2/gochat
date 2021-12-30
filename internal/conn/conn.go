package conn

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/skale2/gochat/internal/db"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var globalCtx context.Context

func Initialize(ctx context.Context) {
	globalCtx = ctx
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

func Finalize() {
	pool.cancelAll()
}

func Create(rw http.ResponseWriter, r *http.Request, username model.Username) (chan<- *model.Message, error) {
	// Upgrade http connection to websocket
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("connected to user %v at address %v", username, conn.RemoteAddr())

	// Ensure following reader/writer goroutines are closed when websocket conn is closed
	ctx, cancel := context.WithCancel(globalCtx)
	conn.SetCloseHandler(createCloseHandler(username))

	// Create goroutines for reading/writing to channel
	writeChan := make(chan *model.Message)
	go reader(ctx, conn, writeChan, username)
	go writer(ctx, conn, writeChan)

	// Add connection to active connection pool
	pool.addUser(username, writeChan, conn, cancel)

	return writeChan, nil
}

func reader(ctx context.Context, conn *websocket.Conn, writeChan <-chan *model.Message, username model.Username) {
	var msg model.Message

	for {
		// Break out of listener loop if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read message from websocket
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Error(err)
			return
		}

		msgBytes, _ := json.Marshal(msg)
		log.Infof("received message from user %v: %v", username, string(msgBytes))

		// Extract message and initialize its fields
		msg.Initialize(username)

		// Validate receipient exists
		_, err = db.GetUser(msg.Receipient)
		if err != nil {
			log.Infof("Unable to find user %v", msg.Receipient)
			return
		}

		// Add message to DB
		db.AddMessage(&msg)

		// If user is connected, send to user immediately
		pool.writeToUser(msg.Receipient, &msg)
	}
}

func writer(ctx context.Context, conn *websocket.Conn, writeChan <-chan *model.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-writeChan:
			conn.WriteJSON(msg)
		}
	}
}

func createCloseHandler(username model.Username) func(int, string) error {
	return func(code int, text string) error {
		log.Infof("disconnected from user %v", username)
		pool.removeUser(username)
		return nil
	}
}
