package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./user_management.db")
	if err != nil {
		panic(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		firstname TEXT,
		lastname TEXT,
		email TEXT,
		avatar TEXT,
		phone TEXT,
		dob TEXT,
		country TEXT,
		city TEXT,
		street_name TEXT,
		street_address TEXT
	);`
	_, err = DB.Exec(createTable)
	if err != nil {
		panic(err)
	}
	fmt.Println("Database initialized")
}
