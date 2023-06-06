package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/token"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	// create simple dummy config
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

// create a new context and add auth token to metadata
func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	// add a access token to the context metadata
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}
	return metadata.NewIncomingContext(context.Background(), md)
}
