package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

var Conn *sql.DB

func Connect() error {
	var err error
	Conn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	return err
}
