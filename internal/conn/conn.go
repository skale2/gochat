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

func Initialize() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

func Create(rw http.ResponseWriter, r *http.Request, username model.Username) (chan<- *model.Message, error) {
	// Upgrade http connection to websocket
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("connected to user %v at address %v", username, conn.RemoteAddr())

	// Ensure following reader/writer goroutines are closed when websocket conn is closed
	ctx, cancel := context.WithCancel(context.Background())
	conn.SetCloseHandler(close(username, cancel))

	// Create goroutines for reading/writing to channel
	writeChan := make(chan *model.Message)
	go reader(ctx, conn, writeChan, username)
	go writer(ctx, conn, writeChan)

	// Add connection to active connection pool
	pool.addUser(username, writeChan)

	return writeChan, nil
}

func reader(ctx context.Context, conn *websocket.Conn, writeChan <-chan *model.Message, username model.Username) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var req model.SocketRequest

			// Read message from websocket
			err := conn.ReadJSON(&req)
			if err != nil {
				log.Error(err)
				return
			}

			requestHandler(username, &req)
		}
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

func requestHandler(username model.Username, req *model.SocketRequest) {
	reqBytes, _ := json.Marshal(req)
	log.Infof("received request from user %v: %v", username, string(reqBytes))

	switch req.RequestType {
	case model.RequestTypeSendMessage:
		// Extract message and calculate + set its id
		msg := req.Payload.(model.Message)
		msg.Initialize(username)

		// Validate receipient exists
		_, err := db.GetUser(msg.Receipient)
		if err != nil {
			log.Infof("Unable to find user %v", msg.Receipient)
			return
		}

		// Add message to DB
		db.AddMessage(&msg)

		// If user is connected, send to user immediately
		writeChan, found := pool.retrieveUser(msg.Receipient)
		if found {
			writeChan <- &msg
		}
	case model.RequestTypeReadThread:
		// Extract message ID
		sender := req.Payload.(model.Username)
		db.ReadThread(username, sender)
	default:
		log.Errorf("unknown request type %v", req.RequestType)
	}
}

func close(username model.Username, cancel context.CancelFunc) func(int, string) error {
	return func(code int, text string) error {
		log.Infof("disconnected from user %v", username)
		cancel()
		pool.removeUser(username)
		return nil
	}
}
