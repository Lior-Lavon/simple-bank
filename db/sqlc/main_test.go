package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // postgres drive
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueriers *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err.Error())
		return
	}

	testQueriers = New(testDB) // from db.go file

	code := m.Run() // Run runs the tests. It returns an exit code to pass to os.Exit.

	os.Exit(code) // start running the unit-test
}
