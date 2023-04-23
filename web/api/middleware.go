package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/liorlavon/simplebank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// authMiddleware is a higer order function that will return the Authentication Middleware function
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {

	// return the annonymouse Authentication Middleware function
	return func(ctx *gin.Context) {

		// get the authorization header from the request
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		// expect form : [Bearer accesstoken]

		// split the autohrizationHeader by space
		fields := strings.Split(authorizationHeader, " ")
		if len(fields) < 2 {
			err := errors.New("invalid autohrization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// token is valid !!

		// store the payload in the context
		ctx.Set(authorizationPayloadKey, payload)

		// call Next() to pass the ctx to the next Middleware or the main handler of the route
		ctx.Next()
	}
}
