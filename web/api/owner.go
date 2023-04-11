package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

func (s *Server) createOwner(ctx *gin.Context) {
	var request struct {
		Firstname string `json:"firstname" binding:"required"`
		Lastname  string `json:"lastname" binding:"required"`
		Email     string `json:"email" binding:"required"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateOwnerParams{
		Firstname: request.Firstname,
		Lastname:  request.Lastname,
		Email:     request.Email,
	}

	owner, err := s.store.CreateOwner(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, owner)
}

func (s *Server) getOwner(ctx *gin.Context) {

	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}

	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	o, err := s.store.GetOwner(ctx, request.ID)
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

func (s *Server) listOwners(ctx *gin.Context) {

	var request struct {
		PageId   int32 `form:"page_id" binding:"required,min=1"`
		PageSize int32 `form:"page_size" binding:"required,min=1,max=10"`
	}

	err := ctx.ShouldBindQuery(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListOwnersParams{
		Limit:  int32(request.PageSize),
		Offset: (request.PageId - 1) * request.PageSize,
	}
	list, err := s.store.ListOwners(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, list)
}

func (s *Server) updateOwner(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var body struct {
		ID        int64  `json:"id"`
		Firstname string `json:"firstname" binding:"required"`
		Lastname  string `json:"lastname" binding:"required"`
		Email     string `json:"email" binding:"required"`
	}
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if request.ID != body.ID {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("id does not match")))
		return
	}

	arg := db.UpdateOwnerParams{
		ID:        request.ID,
		Firstname: body.Firstname,
		Lastname:  body.Lastname,
		Email:     body.Email,
	}

	owner, err := s.store.UpdateOwner(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, owner)
}

func (s *Server) deleteOwner(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.store.DeleteOwner(ctx, request.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := map[string]string{
		"response": fmt.Sprintf("owner %d deleted", request.ID),
	}
	ctx.JSON(http.StatusOK, res)
}
