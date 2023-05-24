package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq" // postgres drive
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/gapi"
	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/web/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	count int64
)

// @title Simple-Bank Service API
func main() {

	// read config
	config, err := util.LoadConfig(".") // path is curent folder
	if err != nil {
		log.Fatal("cannot load configuration: ", err)
		return
	}

	//printConfig(config)

	conn, err := connectToDB(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
		return
	}

	store := db.NewStore(conn)

	// start http server
	//runGinServer(config, store)

	// start the http gateway server
	go runGatewayServer(config, store)
	// start gRPC server
	runGrpcServer(config, store)

}

// Grpc Server
func runGrpcServer(config util.Config, store db.Store) {
	// out own implementation of gRPC
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	// start the server to listen to gRPC on a specific port
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create gRPC listener: ", err)
		return
	}

	log.Printf("Start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server: ", err)
		return
	}
}

// Grpc Server
func runGatewayServer(config util.Config, store db.Store) {
	// out own implementation of gRPC
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
		return
	}

	// enable snake case for the gRPC gateway serverset configuration to use the json naming format
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // this will be executed before this run gateway function, to prevent the system of doing unnececery work

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server: ", err)
	}

	// this mux will receive http request from clients
	mux := http.NewServeMux()
	// to convert them to grpc format, we will reroute them to the grpc mux we created before
	mux.Handle("/", grpcMux)

	// start the server to listen to gRPC on a specific port
	//listener, err := net.Listen("tcp", config.GRPCServerAddress)
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create gRPC listener: ", err)
		return
	}

	log.Printf("Start http gateway server at %s", listener.Addr().String())
	//err = grpcServer.Serve(listener)
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start http gateway server: ", err)
		return
	}
}

// HTTP server gin
func runGinServer(config util.Config, store db.Store) {
	// start the http/gin server
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
		return
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start the server: ", err)
		return
	}
}

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

func printConfig(config util.Config) {

	log.Printf("DB_DRIVER: %s\n", config.DBDriver)
	log.Printf("DB_SOURCE: %s\n", config.DBSource)
	log.Printf("HTTP_SERVER_ADDRESS: %s\n", config.HTTPServerAddress)
	log.Printf("GRPC_SERVER_ADDRESS: %s\n", config.GRPCServerAddress)
	log.Printf("TOKEN_SYMMETRIC_KEY: %s\n", config.TokenSymmetricKey)
	log.Printf("ACCESS_TOKEN_DURATION: %s\n", config.AccessTokenDuration)
}
