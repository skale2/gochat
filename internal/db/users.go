package db

import "github.com/skale2/gochat/internal/model"

var (
	createUsersTable = `
	CREATE TABLE IF NOT EXISTS Users (
		Username text PRIMARY KEY,
		Password text
	)`

	addUserStmt = `
	INSERT INTO Users (Username, Password) 
	VALUES (?, ?)`

	getUserStmt = `
	SELECT Username, Password
	FROM Users
	WHERE Username = ?`
)

func AddUser(user *model.User) error {
	_, err := db.Exec(addUserStmt, user.Username, user.Password)
	return err
}

func GetUser(username model.Username) (*model.User, error) {
	var user model.User
	row := db.QueryRow(getUserStmt, username)

	err := row.Scan(&user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
