package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

// will handle all http request
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// Create a new server , and setup all routes
func NewServer(store *db.Store) *Server {
	// create server
	server := &Server{
		store:  store,
		router: gin.Default(),
	}

	// define routes to the router
	server.router.POST("/api/v1/owners", server.createOwner)
	server.router.GET("/api/v1/owners/:id", server.getOwner)
	server.router.GET("/api/v1/owners", server.listOwners)
	server.router.PUT("/api/v1/owners/:id", server.updateOwner)
	server.router.DELETE("/api/v1/owners/:id", server.deleteOwner)

	return server
}

// Start the http server on the address and listen to API requests
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
