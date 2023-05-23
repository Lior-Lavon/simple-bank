package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/util"
)

type createUserParam struct {
	Username  string `json:"username" binding:"required,alphanum"`
	Password  string `json:"password" binding:"required,min=6"`
	Firstname string `json:"firstname" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	Firstname         string    `json:"firstname"`
	Lastname          string    `json:"lastname"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		Firstname:         user.Firstname,
		Lastname:          user.Lastname,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// PingExample godoc
// @Summary Create user
// @Schemes
// @Description do CreateUser
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {string} json
// @Router /api/v1/users [post]
func (s *Server) createUser(ctx *gin.Context) {
	var request createUserParam
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(request.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.CreateUserParams{
		Username:       request.Username,
		HashedPassword: hashedPassword,
		Firstname:      request.Firstname,
		Lastname:       request.Lastname,
		Email:          request.Email,
	}

	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		// try to convert the error to a err.(*pq.Error) type
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	// create a response to hide the hased_password
	res := newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}

// PingExample godoc
// @Summary Get user
// @Schemes
// @Description do GetUser
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {string} json
// @Router /api/v1/users/$username [get]
func (s *Server) getUser(ctx *gin.Context) {

	var request struct {
		Username string `uri:"username" binding:"required"`
	}

	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	o, err := s.store.GetUser(ctx, request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	res := newUserResponse(o)
	ctx.JSON(http.StatusOK, res)
}

func (s *Server) listUsers(ctx *gin.Context) {

	var request struct {
		PageId   int32 `form:"page_id" binding:"required,min=1"`
		PageSize int32 `form:"page_size" binding:"required,min=1,max=10"`
	}

	err := ctx.ShouldBindQuery(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit:  int32(request.PageSize),
		Offset: (request.PageId - 1) * request.PageSize,
	}
	list, err := s.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, list)
}

type updateUserParam struct {
	Username  string `json:"username" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

func (s *Server) updateUser(ctx *gin.Context) {
	var request struct {
		Username string `uri:"username" binding:"required"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var body updateUserParam
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if request.Username != body.Username {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("username does not match")))
		return
	}

	arg := db.UpdateUserParams{
		Username:  request.Username,
		Firstname: body.Firstname,
		Lastname:  body.Lastname,
		Email:     body.Email,
	}

	user, err := s.store.UpdateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *Server) deleteUser(ctx *gin.Context) {
	var request struct {
		Username string `uri:"username"`
	}

	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.store.DeleteUser(ctx, request.Username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := map[string]string{
		"response": fmt.Sprintf("user %s deleted", request.Username),
	}
	ctx.JSON(http.StatusOK, res)
}

// ////// LOGIN REQUEST
type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"` // ID of the refresh token
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (s *Server) loginUser(ctx *gin.Context) {

	var request loginUserRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get the user
	user, err := s.store.GetUser(ctx, request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// verify password
	err = util.CheckPassword(request.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// create authentication PASETO token
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(request.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(request.Username, s.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	createSessionParams := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     request.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(), // exist in the gin framework
		ClientIp:     ctx.ClientIP(),          // exist in the gin framework
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	session, err := s.store.CreateSession(ctx, createSessionParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create response
	res := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, res)
}
