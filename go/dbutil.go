package friezechat

import (
	"database/sql"
)

type Env struct {
	db *sql.DB
}

var Envdb *Env

func InitPostgres() {
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
