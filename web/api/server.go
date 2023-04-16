package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

// Server struct
type Server struct {
	store  db.Store
	router *gin.Engine
}

// Create a new server , and setup all routes
func NewServer(store db.Store) *Server {
	// create server
	server := &Server{
		store:  store,
		router: gin.Default(),
	}

	// get currect validator engine(interface) and conver it to *validator.Validate pointer
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		// register new validator
		v.RegisterValidation("currency", validCurrency)
	}

	// owners routing
	server.router.POST("/api/v1/owners", server.createOwner)
	server.router.GET("/api/v1/owners/:id", server.getOwner)
	server.router.GET("/api/v1/owners", server.listOwners)
	server.router.PUT("/api/v1/owners/:id", server.updateOwner)
	server.router.DELETE("/api/v1/owners/:id", server.deleteOwner)

	// accounts routing
	server.router.POST("/api/v1/accounts", server.createAccount)
	server.router.GET("/api/v1/accounts/:id", server.getAccount)
	server.router.GET("/api/v1/accounts", server.listAccounts)
	server.router.PUT("/api/v1/accounts/:id", server.updateAccount)
	server.router.DELETE("/api/v1/accounts/:id", server.deleteAccount)

	// Entries routing
	server.router.POST("/api/v1/entries", server.createEntry)
	server.router.GET("/api/v1/entries/:id", server.getEntry)
	server.router.GET("/api/v1/entries", server.listEntries)
	server.router.PUT("/api/v1/entries/:id", server.updateEntry)
	server.router.DELETE("/api/v1/entries/:id", server.deleteEntry)

	// Transfers routing
	server.router.POST("/api/v1/transfers", server.createTransfer)
	server.router.GET("/api/v1/transfers/:id", server.getTransfer)
	server.router.GET("/api/v1/transfers", server.listTransfers)
	server.router.PUT("/api/v1/transfers/:id", server.updateTransfer)
	server.router.DELETE("/api/v1/transfers/:id", server.deleteTransfer)

	return server
}

// Start the http server on the address and listen to API requests
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
