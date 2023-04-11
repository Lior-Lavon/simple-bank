package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

func (s *Server) createEntry(ctx *gin.Context) {
	var request struct {
		AccountID int64 `json:"account_id" binding:"required"`
		Amount    int64 `json:"amount" binding:"required"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if account exist
	_, err = s.store.GetAccount(ctx, request.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			res := map[string]string{
				"error": fmt.Sprintf("account %d does not exist", request.AccountID),
			}
			ctx.JSON(http.StatusNotFound, res)
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
	}

	arg := db.CreateEntryParams{
		AccountID: request.AccountID,
		Amount:    request.Amount,
	}

	entry, err := s.store.CreateEntry(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

func (s *Server) getEntry(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	entry, err := s.store.GetEntry(ctx, request.ID)
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

func (s *Server) listEntries(ctx *gin.Context) {
	var request struct {
		PageId   int32 `form:"page_id" binding:"required,min=1"`
		PageSize int32 `form:"page_size" binding:"required,min=1,max=10"`
	}

	err := ctx.ShouldBindQuery(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit:  int32(request.PageSize),
		Offset: int32((request.PageId - 1) * request.PageSize),
	}
	entries, err := s.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

func (s *Server) updateEntry(ctx *gin.Context) {

	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var body struct {
		ID     int64 `json:"id"`
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

	arg := db.UpdateEntryParams{
		ID:     request.ID,
		Amount: body.Amount,
	}

	entry, err := s.store.UpdateEntry(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)

}

func (s *Server) deleteEntry(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.store.DeleteEntry(ctx, request.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := map[string]string{
		"response": fmt.Sprintf("entry %d deleted", request.ID),
	}
	ctx.JSON(http.StatusOK, res)
}
