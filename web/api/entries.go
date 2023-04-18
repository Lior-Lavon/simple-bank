package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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

	arg := db.CreateEntryParams{
		AccountID: request.AccountID,
		Amount:    request.Amount,
	}

	entry, err := s.store.CreateEntry(ctx, arg)
	if err != nil {
		// try to convert the error to a err.(*pq.Error) type
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

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

	// check for extend uri
	var extends struct {
		Extends bool `form:"extends"`
	}
	err = ctx.ShouldBindQuery(&extends)
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

	var res any
	if extends.Extends {
		e, err := prepareEntryResponse(s, ctx, entry)
		if err != nil {
			ctx.JSON(http.StatusBadGateway, errorResponse(err))
			return
		}
		res = e
	} else {
		res = entry
	}

	ctx.JSON(http.StatusOK, res)
}

func (s *Server) listEntries(ctx *gin.Context) {
	var request struct {
		AccountID int64 `form:"account_id" binding:"required,min=1"`
		PageId    int32 `form:"page_id" binding:"required,min=1"`
		PageSize  int32 `form:"page_size" binding:"required,min=1,max=10"`
		Extends   bool  `form:"extends"`
	}

	err := ctx.ShouldBindQuery(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListEntriesParams{
		AccountID: request.AccountID,
		Limit:     int32(request.PageSize),
		Offset:    int32((request.PageId - 1) * request.PageSize),
	}
	entries, err := s.store.ListEntries(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var res any
	if request.Extends {
		var arr []map[string]any
		for _, acc := range entries {
			extAcc, err := prepareEntryResponse(s, ctx, acc)
			if err != nil {
				ctx.JSON(http.StatusBadGateway, errorResponse(err))
				return
			}
			arr = append(arr, extAcc)
		}
		res = arr
	} else {
		res = entries
	}

	ctx.JSON(http.StatusOK, res)
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

func prepareEntryResponse(s *Server, ctx context.Context, entry db.Entry) (res map[string]any, err error) {

	account, err := s.store.GetAccount(ctx, entry.AccountID)
	if err != nil {
		return
	}

	data, _ := json.Marshal(entry)
	json.Unmarshal(data, &res)
	res["account"] = account

	delete(res, "account_id")
	return
}
