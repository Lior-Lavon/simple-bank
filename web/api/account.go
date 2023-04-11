package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

func (s *Server) createAccount(ctx *gin.Context) {
	var request struct {
		OwnerID  int64  `json:"owner_id" binding:"required"`
		Balance  int64  `json:"balance" binding:"required"`
		Currency string `json:"currency" binding:"required,oneof=USD EUR"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if owner exist
	_, err = s.store.GetOwner(ctx, request.OwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			res := map[string]string{
				"error": fmt.Sprintf("owner %d does not exist", request.OwnerID),
			}
			ctx.JSON(http.StatusNotFound, res)
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
	}

	arg := db.CreateAccountParams{
		OwnerID:  request.OwnerID,
		Balance:  request.Balance,
		Currency: request.Currency,
	}

	account, err := s.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

func (s *Server) getAccount(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := s.store.GetAccount(ctx, request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	owner, err := s.store.GetOwner(ctx, account.OwnerID)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	res := accountToMap(account, owner)

	ctx.JSON(http.StatusOK, res)
}

func (s *Server) listAccounts(ctx *gin.Context) {
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
	accounts, err := s.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var res []map[string]any
	for _, acc := range accounts {
		owner, err := s.store.GetOwner(ctx, acc.OwnerID)
		if err != nil {
			ctx.JSON(http.StatusBadGateway, errorResponse(err))
			return
		}

		extAcc := accountToMap(acc, owner)
		res = append(res, extAcc)
	}

	ctx.JSON(http.StatusOK, res)
}

func (s *Server) updateAccount(ctx *gin.Context) {

	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var body struct {
		ID      int64 `json:"id"`
		Balance int64 `json:"balance" binding:"required"`
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

	arg := db.UpdateAccountParams{
		ID:      request.ID,
		Balance: body.Balance,
	}

	account, err := s.store.UpdateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)

}

func (s *Server) deleteAccount(ctx *gin.Context) {
	var request struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	err := ctx.ShouldBindUri(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = s.store.DeleteAccount(ctx, request.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	res := map[string]string{
		"response": fmt.Sprintf("account %d deleted", request.ID),
	}
	ctx.JSON(http.StatusOK, res)
}

func accountToMap(account db.Account, owner db.Owner) (res map[string]any) {

	data, _ := json.Marshal(account)
	json.Unmarshal(data, &res)
	res["owner"] = owner

	return
}
