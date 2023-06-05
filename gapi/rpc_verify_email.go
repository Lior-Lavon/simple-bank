package gapi

import (
	"context"
	"fmt"

	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	//"github.com/liorlavon/simplebank/worker"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {
		fmt.Println("VerifyEmail : VerifyEmailTx error")
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}

	rsp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	return rsp, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := validation.ValidateEmailId(req.EmailId); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}
	if err := validation.ValidateSecretCode(req.SecretCode); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}
	return
}
