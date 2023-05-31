package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq" // postgres drive
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/gapi"
	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/web/api"
	"github.com/liorlavon/simplebank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	count int64
)

// @title Simple-Bank Service API
func main() {

	// load config
	config, err := util.LoadConfig(".") // path is curent folder
	if err != nil {
		log.Fatal().Msg("cannot load configuration: ")
		return
	}

	if config.Environment == "development" {
		// if 'development' : configure log package to use Pretty output, with colots hand human format
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		// print out json
	}

	printConfig(config)

	conn, err := connectToDB(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to db: ")
		return
	}

	// run db migration here using https://github.com/golang-migrate/migrate
	runDBMigration(config.MIGRATION_URL, config.DBSource)

	store := db.NewStore(conn)

	// Setup Redis Distributor and Processor
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributer(redisOpt)
	go runTaskProcessor(redisOpt, store)

	// start http server
	// runGinServer(config, store)

	// start the http gateway server
	go runGatewayServer(config, store, taskDistributor)
	// start gRPC server
	runGrpcServer(config, store, taskDistributor)
}

// run db migration after loading the configuration
func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal().Msg("failed to run migrate up")
		}
	}

	log.Info().Msg("db migrated succesfully")
}

// background task-processor to handle Redis tasks
func runTaskProcessor(redisOpt asynq.RedisConnOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store)
	log.Info().Msg("start Redis task-processor")

	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

// Grpc Server
func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	// out own implementation of gRPC
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server")
		return
	}

	// create grpc interceptor to log requests
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	// start the server to listen to gRPC on a specific port
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create gRPC listener 1")
		return
	}

	log.Info().Msgf("Start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msg("cannot start gRPC server")
		return
	}
}

// Grpc Server
func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	// out own implementation of gRPC
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server")
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
		log.Fatal().Msg("cannot register handler server")
	}

	// this mux will receive http request from clients
	mux := http.NewServeMux()
	// to convert them to grpc format, we will reroute them to the grpc mux we created before
	mux.Handle("/", grpcMux)

	// start the server to listen to gRPC on a specific port
	//listener, err := net.Listen("tcp", config.GRPCServerAddress)
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create gRPC listener 2")
		return
	}

	log.Info().Msgf("Start http gateway server at %s", listener.Addr().String())
	// set the http logger middleware
	handler := gapi.HttpLoggerMiddleware(mux) // return an http logger middleware
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Msg("cannot start http gateway server")
		return
	}
}

// HTTP server gin
func runGinServer(config util.Config, store db.Store) {
	// start the http/gin server
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server")
		return
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start the server")
		return
	}
}

func connectToDB(driverName, dataSourceName string) (*sql.DB, error) {
	for {
		conn, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Info().Msg("Postgress not yet ready ...")
			count++
		} else {
			log.Info().Msg("Connected to Postgre !!")
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
	log.Info().Str("ENVIROMENT", config.Environment).
		Str("DB_DRIVER", config.DBDriver).
		Str("DB_SOURCE", config.DBSource).
		Str("MIGRATION_URL", config.MIGRATION_URL).
		Str("HTTP_SERVER_ADDRESS", config.HTTPServerAddress).
		Str("GRPC_SERVER_ADDRESS", config.GRPCServerAddress).
		Str("TOKEN_SYMMETRIC_KEY", config.TokenSymmetricKey).
		Str("ACCESS_TOKEN_DURATION", config.AccessTokenDuration.String()).
		Str("REFRESH_TOKEN_DURATION", config.RefreshTokenDuration.String()).
		Msg("Config: ")
}
