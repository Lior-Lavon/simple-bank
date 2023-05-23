package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/token"
	"github.com/liorlavon/simplebank/util"

	// gin-swagger middleware
	_ "github.com/liorlavon/simplebank/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server struct
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// Create a new server , and setup all routes
func NewServer(config util.Config, store db.Store) (*Server, error) {
	// initial tokenMaker
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker : %w", err)
	}

	// create server
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		router:     nil, // set up later
	}

	// get currect validator engine(interface) and conver it to *validator.Validate pointer
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		// register new validator
		v.RegisterValidation("currency", validCurrency)
		//v.RegisterValidation("email", validEmail)
	}

	// create router and add routes
	server.router = server.setupRoute()

	return server, nil
}

func (server *Server) setupRoute() *gin.Engine {
	router := gin.Default()

	// add Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// users routing
	router.POST("/api/v1/users/login", server.loginUser)
	router.POST("/api/v1/users", server.createUser)

	// token renew access
	router.POST("/api/v1/token/renew_access", server.renewAccessToken)

	// add all routes that have common middlewares or the same group router.
	authGroup := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authGroup.GET("/api/v1/users/:username", server.getUser)
	router.GET("/api/v1/users", server.listUsers)
	authGroup.PUT("/api/v1/users/:username", server.updateUser)
	router.DELETE("/api/v1/users/:username", server.deleteUser)

	// accounts routing
	authGroup.POST("/api/v1/accounts", server.createAccount)
	authGroup.GET("/api/v1/accounts/:id", server.getAccount)
	authGroup.GET("/api/v1/accounts", server.listAccounts)
	authGroup.PUT("/api/v1/accounts/:id", server.updateAccount)
	authGroup.DELETE("/api/v1/accounts/:id", server.deleteAccount)

	// Entries routing
	authGroup.POST("/api/v1/entries", server.createEntry)
	authGroup.GET("/api/v1/entries/:id", server.getEntry)
	authGroup.GET("/api/v1/entries", server.listEntries)
	authGroup.PUT("/api/v1/entries/:id", server.updateEntry)
	authGroup.DELETE("/api/v1/entries/:id", server.deleteEntry)

	// Transfers routing
	authGroup.POST("/api/v1/transfers", server.createTransfer)
	authGroup.GET("/api/v1/transfers/:id", server.getTransfer)
	authGroup.GET("/api/v1/transfers", server.listTransfers)
	authGroup.PUT("/api/v1/transfers/:id", server.updateTransfer)
	authGroup.DELETE("/api/v1/transfers/:id", server.deleteTransfer)

	return router
}

// Start the http server on the address and listen to API requests
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
