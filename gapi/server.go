package gapi

import (
	"fmt"

	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/token"
	"github.com/liorlavon/simplebank/util"
)

// Server serve gRPC request for our banking service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

// NewServer creates a new gRPC server
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker : %w", err)
	}

	// create server
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
