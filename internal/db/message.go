package db

import (
	"database/sql"

	"github.com/skale2/gochat/internal/model"
)

var (
	createMessagesTable = `
	CREATE TABLE IF NOT EXISTS Messages (
		Id text PRIMARY KEY,
		Time timestamp,
		Content text,
		Sender text,
		Receipient text,
		Read integer
	)`

	addMessageStmt = `
	INSERT INTO Messages (Id, Time, Content, Sender, Receipient, Read)
	VALUES (?, ?, ?, ?, ?, ?)`

	readThreadStmt = `
	UPDATE Messages 
	SET Read = TRUE 
	WHERE Receipient = ? AND Sender = ?`

	getThreadMessagesStmt = `
	SELECT Time, Content, Sender, Receipient, Read
	FROM Messages 
	WHERE Receipient = ? AND Sender = ?
	ORDER BY Time DESC`

	getThreadsStmt = `
	SELECT MAX(Time), Sender, SUM(Read)
	FROM Messages 
	WHERE Receipient = ?
	GROUP BY Sender
	ORDER BY Time DESC`

	getUnreadMessagesStmt = `
	SELECT Time, Content, Sender, Receipient, Read
	FROM Messages 
	WHERE Receipient = ? AND Read = FALSE
	ORDER BY Time DESC`
)

func AddMessage(msg *model.Message) error {
	_, err := db.Exec(addMessageStmt, msg.Id, msg.Time, msg.Content, msg.Sender, msg.Receipient, msg.Read)
	return err
}

func ReadThread(receipient, sender model.Username) error {
	_, err := db.Exec(readThreadStmt, receipient, sender)
	return err
}

func GetThreads(receipient model.Username) ([]*model.Thread, error) {
	rows, err := db.Query(getThreadsStmt, receipient)
	return retrieveThreadRows(rows, err)
}

func GetThreadMessages(receipient, sender model.Username) ([]*model.Message, error) {
	rows, err := db.Query(getThreadMessagesStmt, receipient, sender)
	return retrieveMessageRows(rows, err)
}

func GetUnreadMessages(receipient model.Username) ([]*model.Message, error) {
	rows, err := db.Query(getUnreadMessagesStmt, receipient)
	return retrieveMessageRows(rows, err)
}

func retrieveThreadRows(rows *sql.Rows, err error) ([]*model.Thread, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []*model.Thread
	for rows.Next() {
		var thread model.Thread

		err := rows.Scan(&thread.Time, &thread.Sender, &thread.Read)
		if err != nil {
			return threads, err
		}

		threads = append(threads, &thread)
	}

	if err = rows.Err(); err != nil {
		return threads, err
	} else {
		return threads, nil
	}
}

func retrieveMessageRows(rows *sql.Rows, err error) ([]*model.Message, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var msg model.Message

		err := rows.Scan(&msg.Id, &msg.Time, &msg.Content, &msg.Sender, &msg.Receipient, &msg.Read)
		if err != nil {
			return messages, err
		}

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return messages, err
	} else {
		return messages, nil
	}
}
