package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	code := m.Run() // Run runs the tests. It returns an exit code to pass to os.Exit.
	os.Exit(code)   // start running the unit-test
}
