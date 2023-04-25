package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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

	// conn, err := connectToDB(config.DBDriver, config.DBSource)
	// if err != nil {
	// 	log.Fatal("cannot connect to db: ", err)
	// 	return
	// }

	conn := connectToDB()
	if conn == nil {
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

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	log.Println("dsn : ", dsn)
	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgress not yet ready ...")
			count++
		} else {
			fmt.Println("Connected to Postgre !!")
			return connection
		}

		if count > 10 {
			//log.Panic("Exit connectionDB !!")
			return nil
		}

		time.Sleep(2 * time.Second)
		//		continue
	}
}

// open the connection to db
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

/*
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
*/
