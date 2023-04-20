package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // postgres drive
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/web/api"
)

func main() {

	// read config
	config, err := util.LoadConfig(".") // path is curent folder
	if err != nil {
		log.Fatal("cannot load configuration: ", err)
		return
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
		return
	}

	s := db.NewStore(conn)
	server, err := api.NewServer(config, s)
	if err != nil {
		log.Fatal("cannot create server: ", err)
		return
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start the server: ", err)
		return
	}
}
