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

var (
	count int64
)

func main() {

	// read config
	config, err := util.LoadConfig(".") // path is curent folder
	if err != nil {
		log.Fatal("cannot load configuration: ", err)
		return
	}

	printConfig(config)

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
	for {
		conn, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Println("Postgress not yet ready ...")
			count++
		} else {
			fmt.Println("Connected to Postgre !!")
			return conn, nil
		}

		if count > 10 {
			//log.Panic("Exit connectionDB !!")
			return nil, fmt.Errorf("failed to connect to db")
		}

		time.Sleep(2 * time.Second)
		//		continue
	}
}

func printConfig(config util.Config) {
	log.Printf("DB_DRIVER: %s\n", config.DBDriver)
	log.Printf("DB_SOURCE: %s\n", config.DBSource)
	log.Printf("SERVER_ADDRESS: %s\n", config.ServerAddress)
	log.Printf("TOKEN_SYMMETRIC_KEY: %s\n", config.TokenSymmetricKey)
	log.Printf("ACCESS_TOKEN_DURATION: %s\n", config.AccessTokenDuration)
}
