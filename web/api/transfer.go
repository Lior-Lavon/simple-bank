package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/token"
)

func (s *Server) createTransfer(ctx *gin.Context) {
	var request struct {
		FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
		ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
		Amount        int64  `json:"amount" binding:"required,gte=1,lte=100"`
		Currency      string `json:"currency" binding:"required,currency"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := s.validAccount(ctx, request.FromAccountID, request.Currency)
	if !valid {
		return
	}

	// Authorization : a user can only transfare money from his own account
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := fmt.Errorf("from account does not belong to the authenticated user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	_, valid = s.validAccount(ctx, request.ToAccountID, request.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountId: request.FromAccountID,
		ToAccountId:   request.ToAccountID,
		Amount:        request.Amount,
	}

	result, err := s.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// check if an account with ID ext and the currency matches as given
func (s *Server) validAccount(ctx *gin.Context, accouID int64, currency string) (db.Account, bool) {
	account, err := s.store.GetAccount(ctx, accouID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Sprintf("account [%v] currency mismatch %s, %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New(err)))
		return account, false
	}

	return account, true
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

	// check for extend uri
	var extends struct {
		Extends bool `form:"extends"`
	}
	err = ctx.ShouldBindQuery(&extends)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	transfer, err := s.store.GetTransfer(ctx, request.ID)
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
		t, err := prepareTransferResponse(s, ctx, transfer)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		res = t
	} else {
		res = transfer
	}

	ctx.JSON(http.StatusOK, res)
}

func (s *Server) listTransfers(ctx *gin.Context) {
	var request struct {
		FromAccountID int64 `form:"from_account_id" binding:"required"`
		PageId        int32 `form:"page_id" binding:"required,min=1"`
		PageSize      int32 `form:"page_size" binding:"required,min=1,max=10"`
		Extends       bool  `form:"extends"`
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
	transfers, err := s.store.ListTransfersFrom(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var res any

	if request.Extends {
		var arr []map[string]any
		for _, e := range transfers {
			resT, err := prepareTransferResponse(s, ctx, e)
			if err != nil {
				ctx.JSON(http.StatusBadGateway, errorResponse(err))
				return
			}

			arr = append(arr, resT)
		}
		res = arr
	} else {
		res = transfers
	}

	ctx.JSON(http.StatusOK, res)
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

func prepareTransferResponse(s *Server, ctx context.Context, transfer db.Transfer) (res map[string]any, err error) {

	fromAccount, err := s.store.GetAccount(ctx, transfer.FromAccountID)
	if err != nil {
		return
	}

	toAccount, err := s.store.GetAccount(ctx, transfer.ToAccountID)
	if err != nil {
		return
	}

	data, _ := json.Marshal(transfer)
	json.Unmarshal(data, &res)
	res["from_account"] = fromAccount
	res["to_account"] = toAccount

	delete(res, "from_account_id")
	delete(res, "to_account_id")

	return
}
