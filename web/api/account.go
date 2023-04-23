package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/token"
)

type createAccountRequest struct {
	//	Owner    string `json:"owner" binding:"required"`
	Balance  int64  `json:"balance" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (s *Server) createAccount(ctx *gin.Context) {
	var request createAccountRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Authorization : a user can only create an account for himself
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username, // add the auth user to the create account param
		Balance:  request.Balance,
		Currency: request.Currency,
	}

	account, err := s.store.CreateAccount(ctx, arg)
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

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getAccount(ctx *gin.Context) {
	var request getAccountRequest
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

	account, err := s.store.GetAccount(ctx, request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Authorization : a user should only be able to get an account that he owns
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	// check if this account.owner belongs to the authPayload.username
	if account.Owner != authPayload.Username {
		err := fmt.Errorf("account does not belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	var res any
	if extends.Extends {
		res, err = prepareAccountResponse(s, ctx, account)
		if err != nil {
			ctx.JSON(http.StatusBadGateway, errorResponse(err))
			return
		}
	} else {
		res = account
	}

	ctx.JSON(http.StatusOK, res)
}

func (s *Server) listAccounts(ctx *gin.Context) {

	var request struct {
		PageId   int32 `form:"page_id" binding:"required,min=1"`
		PageSize int32 `form:"page_size" binding:"required,min=1,max=10"`
		Extends  bool  `form:"extends"`
	}

	err := ctx.ShouldBindQuery(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Authorization : a user can only list accounts that belongs to him
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  int32(request.PageSize),
		Offset: int32((request.PageId - 1) * request.PageSize),
	}

	accounts, err := s.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, errorResponse(err))
		return
	}

	var res any
	if request.Extends {
		arr := []map[string]any{}
		for _, acc := range accounts {
			extAcc, err := prepareAccountResponse(s, ctx, acc)
			if err != nil {
				ctx.JSON(http.StatusBadGateway, errorResponse(err))
				return
			}

			arr = append(arr, extAcc)
		}
		res = arr
	} else {
		res = accounts
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
		ID      int64 `json:"id" binding:"required,min=1"`
		Balance int64 `json:"balance" binding:"required"`
	}
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if request.ID != body.ID {
		err := fmt.Errorf("id %d does not match", request.ID)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
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

func prepareAccountResponse(s *Server, ctx context.Context, account db.Account) (res map[string]any, err error) {

	owner, err := s.store.GetUser(ctx, account.Owner)
	if err != nil {
		return
	}

	data, _ := json.Marshal(account)
	json.Unmarshal(data, &res)
	res["owner"] = owner

	delete(res, "owner_id")
	return
}
