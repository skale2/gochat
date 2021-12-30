package conn

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/model"
)

const (
	SendMessageJob = iota
	ReadMessageJob
)

var pool connPool = connPool{
	conns: make(map[model.Username]connEntry, 0),
}

type connEntry struct {
	WriteChan chan<- *model.Message
	Conn      *websocket.Conn
	Cancel    context.CancelFunc
}

type connPool struct {
	sync.RWMutex
	conns map[model.Username]connEntry
}

func (pool *connPool) writeToUser(username model.Username, msg *model.Message) bool {
	pool.RLock()
	defer pool.RUnlock()

	entry, found := pool.conns[username]
	if !found {
		return false
	}

	entry.WriteChan <- msg
	return true
}

func (pool *connPool) addUser(
	username model.Username,
	writeChan chan *model.Message,
	conn *websocket.Conn,
	cancel context.CancelFunc,
) {
	pool.Lock()
	defer pool.Unlock()

	pool.conns[username] = connEntry{writeChan, conn, cancel}
}

func (pool *connPool) removeUser(username model.Username) bool {
	pool.Lock()
	defer pool.Unlock()

	if entry, exists := pool.conns[username]; exists {
		entry.Cancel()
		delete(pool.conns, username)
		return true
	} else {
		return false
	}
}

func (pool *connPool) cancelAll() {
	pool.Lock()
	defer pool.Unlock()

	for username, entry := range pool.conns {
		entry.Cancel()
		if err := entry.Conn.Close(); err != nil {
			log.Errorf("unable to disconnect from user: %v", err)
			continue
		}

		log.Infof("disconnected from user %v", username)
	}
}
