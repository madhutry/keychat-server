package main

import (
	"database/sql"
	"fmt"
	"os"
)

type Env struct {
	db *sql.DB
}

var Envdb *Env

func InitPostgres() {
	fmt.Println("Getting DBURL")
	fmt.Println(os.Getenv("DB_URL"))
	connString := GetDBUrl()
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	err = db.Ping()

	if err != nil {
		panic(err)
	}
	Envdb = &Env{db: db}
}

/* func InitSqllite() {
	db, err := sql.Open("sqlite3", "/home/gogreek/projects/code/git/frieze-db/chat.db")
	if err != nil {
		log.Fatal(err)
	}
	Envdb = &Env{db: db}
}
*/
