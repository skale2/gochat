package conn

import (
	"sync"

	"github.com/skale2/gochat/internal/model"
)

const (
	SendMessageJob = iota
	ReadMessageJob
)

var pool connPool = connPool{
	userConns: make(map[model.Username]chan<- *model.Message, 0),
}

type connPool struct {
	sync.RWMutex
	userConns map[model.Username]chan<- *model.Message
}

func (pool *connPool) retrieveUser(username model.Username) (chan<- *model.Message, bool) {
	pool.RLock()
	defer pool.RUnlock()

	writeChan, found := pool.userConns[username]
	if !found {
		return nil, false
	} else {
		return writeChan, true
	}
}

func (pool *connPool) addUser(username model.Username, writeChan chan *model.Message) {
	pool.Lock()
	defer pool.Unlock()

	pool.userConns[username] = writeChan
}

func (pool *connPool) removeUser(username model.Username) bool {
	pool.Lock()
	defer pool.Unlock()

	if _, exists := pool.userConns[username]; exists {
		delete(pool.userConns, username)
		return true
	} else {
		return false
	}
}
