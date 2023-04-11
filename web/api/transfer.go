package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

func (s *Server) createTransfer(ctx *gin.Context) {
	var request struct {
		FromAccountID int64 `json:"from_account_id" binding:"required"`
		ToAccountID   int64 `json:"to_account_id" binding:"required"`
		Amount        int64 `json:"amount" binding:"required"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if account exist
	_, err = s.store.GetAccount(ctx, request.FromAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			res := map[string]string{
				"error": fmt.Sprintf("account %d does not exist", request.FromAccountID),
			}
			ctx.JSON(http.StatusNotFound, res)
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
	}

	_, err = s.store.GetAccount(ctx, request.ToAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			res := map[string]string{
				"error": fmt.Sprintf("account %d does not exist", request.ToAccountID),
			}
			ctx.JSON(http.StatusNotFound, res)
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
	}

	arg := db.CreateTransferParams{
		FromAccountID: request.FromAccountID,
		ToAccountID:   request.ToAccountID,
		Amount:        request.Amount,
	}

	transfer, err := s.store.CreateTransfer(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

func (s *Server) getTransfer(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	entry, err := s.store.GetTransfer(ctx, request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

func (s *Server) listTransfers(ctx *gin.Context) {
	var request struct {
		FromAccountID int64 `form:"from_account_id" binding:"required"`
		PageId        int32 `form:"page_id" binding:"required,min=1"`
		PageSize      int32 `form:"page_size" binding:"required,min=1,max=10"`
	}

	err := ctx.ShouldBindQuery(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListTransfersFromParams{
		FromAccountID: request.FromAccountID,
		Limit:         int32(request.PageSize),
		Offset:        int32((request.PageId - 1) * request.PageSize),
	}
	entries, err := s.store.ListTransfersFrom(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

func (s *Server) updateTransfer(ctx *gin.Context) {

	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var body struct {
		ID     int64 `json:"id" binding:"required"`
		Amount int64 `json:"amount" binding:"required"`
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

	arg := db.UpdateTransferParams{
		ID:     request.ID,
		Amount: body.Amount,
	}

	entry, err := s.store.UpdateTransfer(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)

}

func (s *Server) deleteTransfer(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.store.DeleteTransfer(ctx, request.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := map[string]string{
		"response": fmt.Sprintf("transfer %d deleted", request.ID),
	}
	ctx.JSON(http.StatusOK, res)
}
