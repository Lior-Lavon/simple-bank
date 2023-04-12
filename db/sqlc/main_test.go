package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // postgres drive
	"github.com/liorlavon/simplebank/util"
)

var testQueriers *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {

	// load config file
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config file")
		return
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err.Error())
		return
	}

	testQueriers = New(testDB) // from db.go file

	code := m.Run() // Run runs the tests. It returns an exit code to pass to os.Exit.

	os.Exit(code) // start running the unit-test
}
