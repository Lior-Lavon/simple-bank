package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/util"
)

type createUserParam struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

func (s *Server) createUser(ctx *gin.Context) {
	var request createUserParam
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hp, _ := util.HashPassword(request.Password)
	arg := db.CreateUserParams{
		Username:       request.Username,
		HashedPassword: hp,
		Firstname:      request.Firstname,
		Lastname:       request.Lastname,
		Email:          request.Email,
	}

	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

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

	ctx.JSON(http.StatusOK, o)
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

	fmt.Printf("\n userName :%+v\n", request.Username)

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
