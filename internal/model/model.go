package model

import (
	"crypto/md5"
	"encoding/json"
	"time"
)

type (
	SocketRequestType int
	Username          string
	MessageId         string
)

const (
	RequestTypeSendMessage SocketRequestType = iota
	RequestTypeReadThread
)

type SocketRequest struct {
	RequestType SocketRequestType `json:"request_type"`
	Payload     interface{}       `json:"payload"`
}

func (r *SocketRequest) UnmarshalJSON(b []byte) error {
	var rawReq struct {
		RequestType SocketRequestType `json:"request_type"`
		Payload     *json.RawMessage  `json:"payload"`
	}

	err := json.Unmarshal(b, &rawReq)
	if err != nil {
		return err
	}

	var req SocketRequest = SocketRequest{
		RequestType: rawReq.RequestType,
	}

	switch req.RequestType {
	case RequestTypeSendMessage:
		var msg Message

		err = json.Unmarshal(*rawReq.Payload, &msg)
		if err != nil {
			return err
		}

		req.Payload = msg

	case RequestTypeReadThread:
		var sender Username

		err = json.Unmarshal(*rawReq.Payload, &sender)
		if err != nil {
			return err
		}

		req.Payload = sender
	}

	*r = req
	return nil
}

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
