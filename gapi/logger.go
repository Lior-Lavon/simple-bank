package gapi

import (
	"context"
	"time"

	// used by zerolog
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()

	result, err := handler(ctx, req) // pass to the handler of the request
	duration := time.Since(startTime)

	// get the status code of the request
	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	// get a handler to the log
	logger := log.Info()
	if err != nil {
		// if there is an err from the method, change the sevirity
		logger = log.Error().Err(err)
	}

	logger.Str("protocol", "grpc").
		Str("method", info.FullMethod).      // print method
		Int("status_code", int(statusCode)). // print status code
		Str("status_text", statusCode.String()).
		Dur("duration", duration). // print run duration
		Msg("received a grpc request")

	return result, err
}
