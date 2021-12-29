package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/skale2/gochat/internal/log"
)

const dbFileName = "./database.db"

var db *sql.DB

func Initialize() {
	var err error

	db, err = sql.Open("sqlite3", dbFileName)
	if err != nil {
		log.Error(err)
		return
	}

	_, err = db.Exec(createUsersTable)
	if err != nil {
		log.Error(err)
		return
	}

	_, err = db.Exec(createMessagesTable)
	if err != nil {
		log.Error(err)
		return
	}
}

func Finalize() {
	db.Close()
}
