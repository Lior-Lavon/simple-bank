package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

type createOwnerRequest struct {
	Firstname string `json:"firstname" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
	Email     string `json:"email" binding:"required"`
	// binding:"required,oneof=USD EUR"`
}

func (s *Server) createOwner(ctx *gin.Context) {
	var request createOwnerRequest
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

type getOwnerRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getOwner(ctx *gin.Context) {

	var request getOwnerRequest
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

// type listOwnerRequest struct {
// 	PageId   int64 ``
// 	PageSize int64 ` `
// }

// func (s *Server) listOwners(ctx *gin.Context) {

// }
