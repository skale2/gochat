package model

import (
	"crypto/md5"
	"time"
)

type (
	Username  string
	MessageId string
)

type User struct {
	Username Username `json:"username"`
	Password string   `json:"password,omitempty"`
}

type Thread struct {
	Time   time.Time `json:"time"`
	Sender Username  `json:"sender"`
	Read   bool      `json:"read"`
}

type Message struct {
	Id         MessageId `json:"id,omitempty"`
	Time       time.Time `json:"time,omitempty"`
	Content    string    `json:"content"`
	Receipient Username  `json:"receipient"`
	Sender     Username  `json:"sender,omitempty"`
	Read       bool      `json:"read,omitempty"`
}

func (m *Message) Initialize(sender Username) {
	m.Sender = sender
	m.Time = time.Now()

	bytes := md5.Sum([]byte(m.Time.String() + m.Content + string(m.Sender) + string(m.Receipient)))
	m.Id = MessageId(bytes[:])
}
