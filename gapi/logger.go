package gapi

import (
	"context"
	"net/http"
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

// implement the ResponseWriter as interface and override the WriteHeader function to get the StatusCode
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// implement ResponseWriter interface
func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode) // call the original function to continue the process
}

// implement ResponseWriter interface
func (rec *ResponseRecorder) Write(b []byte) (int, error) {
	rec.Body = b                       // in case of an error, get it from the body
	return rec.ResponseWriter.Write(b) // call the original function to continue the process
}

// Http Logger middleware, to be added to the gRPC Gateway server
func HttpLoggerMiddleware(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		// pass the request to the main handler
		rec := &ResponseRecorder{
			ResponseWriter: res,           // set the original ResponseWriter
			StatusCode:     http.StatusOK, // set the default value, will be updated after the handler.ServeHTTP process
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.Body)
		}

		logger.Str("protocol", "http").
			Str("method", req.Method).                           // print method PUT/GET/POST ...
			Str("path", req.RequestURI).                         // print the full path
			Int("status_code", rec.StatusCode).                  // print status code
			Str("status_text", http.StatusText(rec.StatusCode)). // print the description
			Dur("duration", duration).                           // print run duration
			Msg("received a HTTP request")

	})
}
