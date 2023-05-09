package api

// support refresh token api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ////// RENEW ACCESS TOKEN REQUEST
type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) renewAccessToken(ctx *gin.Context) {

	var request renewAccessTokenRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// verify if the refresh token is valid
	refreshPayload, err := s.tokenMaker.VerifyToken(request.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// get the session
	session, err := s.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if the session is blocked
	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// check if the session.user is the same name as in the refreshPayload
	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// check if the session.refreshToken is the same as req.refreshToken
	if session.RefreshToken != request.RefreshToken {
		err := fmt.Errorf("mismatch session token")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// check if the session expired
	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// create authentication PASETO token
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create response
	res := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, res)
}
