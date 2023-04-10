package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // postgres drive
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/web/api"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = ":8080"
)

func main() {

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err.Error())
		return
	}
	s := db.NewStore(conn)
	server := api.NewServer(s)
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start the server: ", err.Error())
		return
	}
}
