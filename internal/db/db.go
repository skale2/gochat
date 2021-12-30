package db

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/skale2/gochat/internal/log"
)

var db *sql.DB

func SQLLiteDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func Initialize(dbInit *sql.DB) {
	db = dbInit

	_, err := db.Exec(createUsersTable)
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

type ListOptsSource interface {
	Get(key string) string
}

func GetListOpts(src ListOptsSource) (limit, offset int, err error) {
	limit, err = extractListOpt(src, "limit", 1, 15)
	if err == nil {
		offset, err = extractListOpt(src, "offset", 1, 15)
	}

	return
}

func extractListOpt(src ListOptsSource, name string, minVal, defaultVal int) (int, error) {
	s := src.Get(name)
	if len(s) <= 0 {
		return defaultVal, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	} else if i < minVal {
		return 0, fmt.Errorf("%v must be greater than or equal to 0", name)
	}

	return i, nil
}
