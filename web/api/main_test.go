package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	code := m.Run() // Run runs the tests. It returns an exit code to pass to os.Exit.
	os.Exit(code)   // start running the unit-test
}

// create test-server with config
func newTestServer(t *testing.T, store db.Store) *Server {
	// create simple dummy config
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}
