package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

	conn, err := connectToDB(config.DBDriver, config.DBSource)
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

func connectToDB(driverName, dataSourceName string) (*sql.DB, error) {

	counter := 0

	for {
		fmt.Printf("counter = %d\n", counter)
		if counter > 10 {
			return nil, fmt.Errorf("failed to connect to db")
		}
		conn, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Fatal("failed to connect, try again .. ")
			counter++
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("connect succesfull returning")
		return conn, nil
	}
}
